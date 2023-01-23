package model

type User struct {
	MongoID string   `json:"_id,omitempty"     bson:"_id,omitempty"`
	Uid     int      `json:"uid"               bson:"_uid"`
	Name    string   `json:"name"              bson:"name"`
	Age     int      `json:"age"               bson:"age"`
	Friends []string `json:"friends"           bson:"friends"`
}

type UID struct {
	Seq int `bson:"seq"`
}

type FriendsRequest struct {
	Source_id int `json:"source_Id"  bson:"source_id"`
	Target_id int `json:"target_Id"  bson:"target_id"`
}

type RefactorAge struct {
	Name string `json:"name"  bson:"name"`
	Age  int    `json:"age"  bson:"age"`
}

type ShowFriends struct {
	Name string `json:"name"  bson:"name"`
}

type CreateUserDTO struct {
	ID      int      `json:"id"                bson:"_id,omitempty"`
	Name    string   `json:"name"              bson:"name"`
	Age     int      `json:"age"               bson:"age"`
	Friends []string `json:"friends,omitempty" bson:"friends"`
}
