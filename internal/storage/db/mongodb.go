package db

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"httpServer/internal/model"
	"httpServer/pkg/client/mongodb"
	"log"
)

type DB struct {
	collection *mongo.Collection
}

type MongoDB struct {
	host     string `json:"host"     bson:"host"`
	port     string `json:"port"     bson:"port"`
	username string `json:"username" bson:"username"`
	password string `json:"password" bson:"password"`
	database string `json:"database" bson:"database"`
	authDB   string `json:"auth_db"  bson:"auth_db"`
}

type Mongo interface {
	Create(ctx context.Context, user *model.User) (string, error)
	FindOne(ctx context.Context, uid int) (*model.User, error)
	FindAll(ctx context.Context) (u []model.User, err error)
	Update(ctx context.Context, user *model.User, userID int) error
	Delete(ctx context.Context, user *model.User) error
	MakeFriend(ctx context.Context, req model.FriendsRequest) (sourceName, targetName *model.User, err error)
}

func MongoConfig() *MongoDB {
	var TestDB = MongoDB{
		//"localhost",
		"172.17.0.1",
		"27017",
		"",
		"",
		"users",
		""}
	return &TestDB
}

func NewClientMongo(TestDB *MongoDB) (mongoDBClient *mongo.Database, err error) {
	mongoDBClient, err = mongodb.NewClient(context.Background(),
		TestDB.host, TestDB.port,
		TestDB.username, TestDB.password,
		TestDB.database, TestDB.authDB)

	if err != nil {
		log.Print("panic! error")
	}
	return mongoDBClient, nil

}

func NewMongoStorage(mongoDBClient *mongo.Database, collection string) Mongo {
	dbMongo := NewStorageDB(mongoDBClient, collection)
	return dbMongo
}

// InitMongo func initialization
func InitMongo() Mongo {
	TestDB := MongoConfig()
	mongoDBClient, err := NewClientMongo(TestDB)
	if err != nil {
		fmt.Errorf("we have error in func init %v", err)
	}
	dbMongo := NewMongoStorage(mongoDBClient, TestDB.database)
	log.Print("db Mongo is Active now")
	return dbMongo
}

// Create with MongoDB
func (d *DB) Create(ctx context.Context, user *model.User) (string, error) {

	log.Print("[MONGO] create user")
	user.Friends = append(user.Friends, "his friends:")

	uid, err := d.getNext()
	if err != nil {
		log.Print("user.UID.Seq err")
	}
	user.Uid = uid

	result, err := d.collection.InsertOne(nil, user)
	if err != nil {
		log.Print(err)
		return "", err
	}
	log.Print("[MONGO] convert InsertId to ObjectID")
	oid, ok := result.InsertedID.(primitive.ObjectID)
	if ok {
		return oid.Hex(), nil
	}
	log.Print("err")
	return "", nil
}

func (d *DB) getNext() (int, error) {
	upsert := true
	after := options.After
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}

	singleResult := d.collection.FindOneAndUpdate(
		context.TODO(),
		bson.M{"id": "uid"},
		bson.M{"$inc": bson.M{"seq": 1}},
		&opt,
	)

	if singleResult.Err() != nil {
		log.Print("singleResult.Err")
		return 0, singleResult.Err()
	}

	var uid model.UID
	err := singleResult.Decode(&uid)
	if err != nil {
		log.Print("err")
		return 0, err
	}
	return uid.Seq, nil
}

// FindAll with MongoDB
func (d *DB) FindAll(ctx context.Context) (u []model.User, err error) {

	Cursor, err := d.collection.Find(nil, bson.M{})
	if Cursor.Err() != nil {
		log.Print("[MONGO] failed find users out func FindAll")
	}

	if err = Cursor.All(nil, &u); err != nil {
		return u, fmt.Errorf("[MONGO] failed to read from cursor %v", err)
	}

	return u, nil
}

type GotestFind struct {
	uid string
}

// FindOne with MongoDB
func (d *DB) FindOne(ctx context.Context, id int) (u *model.User, err error) {

	filter := bson.M{"_uid": id}

	err = d.collection.FindOne(ctx, filter).Decode(&u)
	if err != nil {
		log.Print("err in decode", err)

		return u, err
	}

	log.Printf("user %v find success ", u)
	return u, nil
}

// Update with MongoDB
func (d *DB) Update(ctx context.Context, u *model.User, userID int) error {

	filter := bson.M{"_uid": userID}
	userBytes, err := bson.Marshal(u)
	if err != nil {
		log.Print(err)
		return fmt.Errorf("error Marshal user")
	}
	var updateUserObj bson.M
	err = bson.Unmarshal(userBytes, &updateUserObj)
	delete(updateUserObj, "_uid")

	update := bson.M{
		"$set": updateUserObj,
	}
	result, err := d.collection.UpdateOne(nil, filter, update)
	if err != nil {
		log.Print(err)
		return fmt.Errorf("failed update")
	}

	if result.MatchedCount == 0 {
		//todo NotFound
		log.Print(err)
		return fmt.Errorf("not found")
	}

	log.Print("update complite")
	return nil
}

// Delete with MongoDB
func (d *DB) Delete(ctx context.Context, u *model.User) error {
	filter := bson.M{"_uid": u.Uid}

	result, err := d.collection.DeleteOne(nil, filter)

	if err != nil {
		log.Print(err)
		return fmt.Errorf("failed to execute query")
	}
	if result.DeletedCount == 0 {
		log.Print("delete count = nil")
		//todo NotFound
		log.Print(err)
		return fmt.Errorf("not found")
	}

	log.Print("delete sucsses")
	return nil
}

// MakeFriend with MongoDB
func (d *DB) MakeFriend(ctx context.Context, req model.FriendsRequest) (sourceName, targetName *model.User, err error) {

	sourceName, targetName, err = d.AllOperationsForFind(ctx, req)

	err = d.FriendlistUpdate(ctx, sourceName, targetName.Name)
	if err != nil {
		fmt.Errorf("error in update friendlist SourceUser")
	}
	err = d.FriendlistUpdate(ctx, targetName, sourceName.Name)
	if err != nil {
		fmt.Errorf("error in update friendlist TargetUser")
	}

	log.Print("update complite")

	return sourceName, targetName, nil
}

// AllOperationsFromSourceUser
func (d *DB) AllOperationsForFind(ctx context.Context, req model.FriendsRequest) (sourceName, targetName *model.User, err error) {
	sourceName, err = d.FindOne(nil, req.Source_id)
	if err != nil {
		log.Print("err in find SourceID with MongoDB")
		return nil, nil, fmt.Errorf("err in find SourceID with MongoDB")
	}

	targetName, err = d.FindOne(nil, req.Target_id)
	if err != nil {
		log.Print("err in find TargetID with MongoDB")
		return nil, nil, fmt.Errorf("err in find TargetID with MongoDB")
	}

	return sourceName, targetName, nil
}

// AllOperationsFromTargetUser
func (d *DB) FriendlistUpdate(ctx context.Context, u *model.User, friendName string) error {

	_, err := d.collection.UpdateOne(nil,
		bson.M{"uid": u.Uid},
		bson.M{"$push": bson.M{"friends": friendName}})
	if err != nil {
		log.Print(err)
		return fmt.Errorf("failed update")
	}

	return nil
}

// NewStorageDB activate Mongo
func NewStorageDB(database *mongo.Database, collection string) Mongo {
	return &DB{
		collection: database.Collection(collection),
	}
}
