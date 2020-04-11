package main

import (
	"context"
	"time"

	"github.com/brianvoe/sjwt"
)

// GetUserByUsername allows you to get a user by their ID
func GetUserByUsername(username string) User {
	var dbUser User
	collection := client.Database("mapdb").Collection("users")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	err := collection.FindOne(ctx, User{Username: username}).Decode(&dbUser)
	if err != nil {
		return User{}
	}
	return dbUser
}

// GetUsernameFromToken will allow us to grab the Username from the token
func GetUsernameFromToken(token string) string {
	if VerifyToken(token) == false {
		return "unidentified"
	}
	claims, _ := sjwt.Parse(token)
	username, _ := claims.GetStr("username")
	return username
}

// GetUserIDFromToken will allow us to grab the UserID from the token
func GetUserIDFromToken(token string) string {
	if VerifyToken(token) == false {
		return "unidentified"
	}
	claims, _ := sjwt.Parse(token)
	userid, _ := claims.GetStr("id")
	return userid
}
