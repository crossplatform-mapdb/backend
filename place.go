package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// CreatePlaceEndpoint allows for you to create a place
func CreatePlaceEndpoint(response http.ResponseWriter, request *http.Request) {
	var token Token
	token.Token = request.Header.Get("token")
	if VerifyToken(token.Token) == false {
		response.WriteHeader(http.StatusUnauthorized)
		response.Write([]byte(`{ "message": "You need to provide a token in the body inorder to access this resource." }`))
		return
	}
	response.Header().Add("content-type", "application/json")
	var place Place
	place.UserID = GetUserIDFromToken(token.Token)
	json.NewDecoder(request.Body).Decode(&place)
	collection := client.Database("mapdb").Collection("places")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	result, _ := collection.InsertOne(ctx, place)
	json.NewEncoder(response).Encode(result)
}

// GetPlacesEndpoint is a function designed to respond with all the places a user has access to.
func GetPlacesEndpoint(response http.ResponseWriter, request *http.Request) {
	var token Token
	token.Token = request.Header.Get("token")
	if VerifyToken(token.Token) == false {
		response.WriteHeader(http.StatusUnauthorized)
		response.Write([]byte(`{ "message": "You need to provide a token in the body inorder to access this resource." }`))
		return
	}
	response.Header().Add("content-type", "application/json")
	var places []Place
	collection := client.Database("mapdb").Collection("users")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(` { "message": "` + err.Error() + `"}`))
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var place Place
		cursor.Decode(&place)
		if place.UserID == GetUserIDFromToken(token.Token) {
			places = append(places, place)
		}
	}
	if err := cursor.Err(); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(` { "message": "` + err.Error() + `"}`))
		return
	}
	json.NewEncoder(response).Encode(places)
}
