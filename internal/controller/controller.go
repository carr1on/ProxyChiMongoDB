package controller

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"httpServer/internal/model"
	"httpServer/internal/storage/db"
	"io"
	"log"
	"net/http"
)

var mongodb db.Mongo

type Server struct {
	Router *chi.Mux
}

type UserResponse struct {
	Status   int        `json:"status"`
	Response model.User `json:"response,omitempty"`
	Error    error      `json:"error,omitempty"`
}

type ResponseDelete struct {
	Status   int    `json:"status"`
	Response string `json:"user,omitempty"`
	Error    error  `json:"error,omitempty"`
}

type UsersResponse struct {
	Status   int          `json:"status"`
	Response []model.User `json:"response,omitempty"`
	Error    error        `json:"error,omitempty"`
}

type UsersFriendResponse struct {
	Status   int      `json:"status"`
	Response []string `json:"response,omitempty"`
	Error    error    `json:"error,omitempty"`
}

func init() {
	mongodb = db.InitMongo()
}

func CreateNewServer() *Server {
	s := &Server{}
	s.Router = chi.NewRouter()
	return s
}

func (s *Server) MountHandlers() {
	r := s.Router
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.AllowContentType("application/json"))
	r.Use(middleware.CleanPath)

	r.Route("/", func(r chi.Router) {
		r.Get("/", DefaultRoute)

		// sub-route 'users'
		r.Route("/user", func(r chi.Router) {
			r.Get("/", GetUser)
			r.Get("/all", GetAllUsers)
			r.Post("/create", CreateUser)
			r.Put("/update", Put)

			r.Post("/make_friend", MakeFriend)
			r.Delete("/delete", DeleteUser)
			r.Route("/this_user", func(r chi.Router) {
				r.Get("/friends", GetFriends)
			})
		})

	})

}

func DefaultRoute(w http.ResponseWriter, _ *http.Request) {
	if _, err := w.Write([]byte("server alive")); err != nil {
		log.Printf("[SERVER] can't send response: %v\n", err)
	}
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var ctx = context.Background()
	content, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		sendResponse(w, http.StatusInternalServerError, &model.User{}, err)
		return
	}

	var u *model.User

	if err = json.Unmarshal(content, &u); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		sendResponse(w, http.StatusInternalServerError, &model.User{}, err)
		return
	}

	if whatCreated, err := mongodb.Create(ctx, u); err != nil {
		log.Printf("error %v in mongo.create user %v", err, whatCreated)
		w.WriteHeader(http.StatusInternalServerError)
		sendResponse(w, http.StatusInternalServerError, u, err)
		log.Printf("[SERVER] error saving user to storage: %v\n", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	sendResponse(w, http.StatusCreated, u, nil)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	var ctx = context.Background()
	content, err := io.ReadAll(r.Body)

	var whatFind *model.User

	var u *model.User
	if err = json.Unmarshal(content, &whatFind); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		sendResponse(w, http.StatusInternalServerError, &model.User{}, err)
		return
	}

	u, err = mongodb.FindOne(ctx, whatFind.Uid)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		err = errors.New("user " + whatFind.Name + " not found")
		sendResponse(w, http.StatusNotFound, &model.User{}, err)
		return
	}

	w.WriteHeader(http.StatusOK)

	sendResponse(w, http.StatusOK, u, nil)
}

func GetAllUsers(w http.ResponseWriter, _ *http.Request) {

	var err error
	var whatFound []model.User
	var ctx = context.Background()
	if whatFound, err = mongodb.FindAll(ctx); err != nil {
		if whatFound == nil {
			w.WriteHeader(http.StatusNotFound)
			err := errors.New("users not found")
			sendResponse(w, http.StatusNotFound, &model.User{}, err)
			return
		}
	}

	w.WriteHeader(http.StatusOK)

	sendResponseSlice(w, http.StatusOK, whatFound, nil)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {

	var ctx = context.Background()

	content, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		sendResponse(w, http.StatusInternalServerError, &model.User{}, err)
		return
	}

	var u *model.User
	if err = json.Unmarshal(content, &u); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		sendResponse(w, http.StatusInternalServerError, &model.User{}, err)
		return
	}

	err = mongodb.Delete(ctx, u)
	log.Print("u = ", u)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("err inside delete with findOne func %v", err)
		sendResponse(w, http.StatusInternalServerError, &model.User{}, err)
		log.Printf("err inside delete func %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	sendResponseDelete(w, http.StatusOK, u.Name, nil)
	return
}

func Put(w http.ResponseWriter, r *http.Request) {
	content, err := io.ReadAll(r.Body)
	var ctx = context.Background()

	var u *model.User
	if err = json.Unmarshal(content, &u); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = mongodb.Update(ctx, u, u.Uid)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		sendResponse(w, http.StatusInternalServerError, u, err)
		return
	}

	w.WriteHeader(http.StatusOK)

	sendResponse(w, http.StatusOK, u, nil)
}

func MakeFriend(w http.ResponseWriter, r *http.Request) {
	content, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var req model.FriendsRequest
	err = json.Unmarshal(content, &req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		sendResponse(w, http.StatusInternalServerError, &model.User{}, err)
		return
	}

	sourceName, targetName, err := mongodb.MakeFriend(nil, req)
	log.Print(sourceName, " = sourceName")
	log.Print(targetName, " = targetName")

	//err = storage.ForMakeFriend(sourceName, targetName)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		sendResponse(w, http.StatusBadRequest, &model.User{}, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(targetName.Name + " and " + sourceName.Name + " now friends\n"))
	return

}

func GetFriends(w http.ResponseWriter, r *http.Request) {
	var ctx = context.Background()
	content, err := io.ReadAll(r.Body)

	var whatFind *model.User

	var u *model.User
	if err = json.Unmarshal(content, &whatFind); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		sendResponse(w, http.StatusInternalServerError, &model.User{}, err)
		return
	}

	u, err = mongodb.FindOne(ctx, whatFind.Uid)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		err := errors.New("user " + whatFind.Name + " not found")
		sendResponse(w, http.StatusNotFound, &model.User{}, err)
		return
	}

	var friends []string
	for _, user := range u.Friends {
		friends = append(friends, user)
	}

	w.WriteHeader(http.StatusOK)
	sendResponseSliceString(w, http.StatusOK, friends[1:], nil)
}

func sendResponse(w http.ResponseWriter, status int, user *model.User, err error) {
	response := UserResponse{
		Status:   status,
		Response: *user,
		Error:    err,
	}
	jsonData, err := json.Marshal(response)
	if err != nil {
		log.Printf("[SERVER] can't prepare response: %v\n", err)
		return
	}
	if _, err = w.Write(jsonData); err != nil {
		log.Printf("[SERVER] can't send response: %v\n", err)
		return
	}
}

func sendResponseDelete(w http.ResponseWriter, status int, user string, err error) {
	response := ResponseDelete{
		Status:   status,
		Response: user,
		Error:    err,
	}
	jsonData, err := json.Marshal(response)
	if err != nil {
		log.Printf("[SERVER] can't prepare response: %v\n", err)
		return
	}
	if _, err := w.Write(jsonData); err != nil {
		log.Printf("[SERVER] can't send response: %v\n", err)
		return
	}
}

func sendResponseSlice(w http.ResponseWriter, status int, users []model.User, err error) {
	response := UsersResponse{
		Status:   status,
		Response: users,
		Error:    err,
	}
	jsonData, err := json.Marshal(response)
	if err != nil {
		log.Printf("[SERVER] can't prepare response: %v\n", err)
		return
	}
	if _, err := w.Write(jsonData); err != nil {
		log.Printf("[SERVER] can't send response: %v\n", err)
		return
	}
}

func sendResponseSliceString(w http.ResponseWriter, status int, friends []string, err error) {
	response := UsersFriendResponse{
		Status:   status,
		Response: friends,
		Error:    err,
	}
	jsonData, err := json.Marshal(response)
	if err != nil {
		log.Printf("[SERVER] can't prepare response: %v\n", err)
		return
	}
	if _, err := w.Write(jsonData); err != nil {
		log.Printf("[SERVER] can't send response: %v\n", err)
		return
	}
}
