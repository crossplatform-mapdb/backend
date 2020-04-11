package main

import "go.mongodb.org/mongo-driver/bson/primitive"

// User defines the structure of a User
type User struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Username string             `json:"username,omitempty" bson:"username,omitempty"`
	Email    string             `json:"email,omitempty" bson:"email,omitempty"`
	Password string             `json:"password,omitempty" bson:"password,omitempty"`
	Admin    string             `json:"admin,omitempty" bson:"admin,omitempty"`
}

// Place defines the structure of a User
type Place struct {
	ID     primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Lon    string             `json:"lon,omitempty" bson:"lon,omitempty"`
	Lat    string             `json:"lat,omitempty" bson:"lat,omitempty"`
	UserID string             `json:"userid,omitempty" bson:"userid,omitempty"`
	Title  string             `json:"title,omitempty" bson:"title,omitempty"`
	Desc   string             `json:"desc,omitempty" bson:"desc,omitempty"`
}

// Token sets up a datamodel for the token
type Token struct {
	Token string `json:"token,omitempty" bson:"token,omitempty"`
}
