package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes the password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash checks to see if the password is correct
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// CreateUserEndpoint Create a new User
func CreateUserEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	var signUpUser User
	var dbUser User
	var db2User User
	json.NewDecoder(request.Body).Decode(&signUpUser)
	collection := client.Database("mapdb").Collection("users")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	dbUser, err := GetUserByUsername(signUpUser.Username)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(` { "message": "` + err.Error() + `"}`))
		return
	}
	if dbUser.Username == signUpUser.Username {
		response.Write([]byte(`{ "message": "` + "An account has already been created with that username, please choose another." + `" }`))
		return
	}
	db2User, err2 := GetUserByEmail(signUpUser.Email)
	if err2 != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(` { "message": "` + err2.Error() + `"}`))
		return
	}
	if db2User.Email == signUpUser.Email {
		response.Write([]byte(`{ "message": "` + "An account has already been created with that email, try logging in." + `" }`))
		return
	}
	signUpUser.Password, _ = HashPassword(signUpUser.Password)
	signUpUser.Admin = "false"
	result, _ := collection.InsertOne(ctx, signUpUser)
	json.NewEncoder(response).Encode(result)
}

// GetUsersEndpoint Get all Users
func GetUsersEndpoint(response http.ResponseWriter, request *http.Request) {
	var token Token
	token.Token = request.Header.Get("token")
	if VerifyToken(token.Token) == false {
		response.WriteHeader(http.StatusUnauthorized)
		response.Write([]byte(`{ "message": "You need to provide a token in the body inorder to access this resource." }`))
		return
	}
	response.Header().Add("content-type", "application/json")
	var users []User
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
		var user User
		cursor.Decode(&user)
		user2, err := GetUserByUsername(GetUsernameFromToken(token.Token))
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
			return
		}
		admin, _ := strconv.ParseBool(user2.Admin)
		if admin == false {
			user.Password = "hidden"
		}
		users = append(users, user)
	}
	if err := cursor.Err(); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(` { "message": "` + err.Error() + `"}`))
		return
	}
	json.NewEncoder(response).Encode(users)
}

// GetUserEndpoint Get a single User
func GetUserEndpoint(response http.ResponseWriter, request *http.Request) {
	var token Token
	token.Token = request.Header.Get("token")
	if VerifyToken(token.Token) == false {
		response.WriteHeader(http.StatusUnauthorized)
		response.Write([]byte(`{ "message": "You need to provide a token in the body inorder to access this resource." }`))
		return
	}
	response.Header().Set("content-type", "application/json")
	params := mux.Vars(request)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var user User
	user, err := GetUserByID(id)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	admin, _ := GetUserByUsername(GetUsernameFromToken(token.Token))
	adminCheck, _ := strconv.ParseBool(admin.Admin)
	if adminCheck == false {
		user.Password = "hidden"
	}
	json.NewEncoder(response).Encode(user)
}

// LoginEndpoint checks to see if a user is who they claim to be and sends them a JWT.
func LoginEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	var loginUser User
	var dbUser User
	json.NewDecoder(request.Body).Decode(&loginUser)
	dbUser, err := GetUserByUsername(loginUser.Username)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	match := CheckPasswordHash(loginUser.Password, dbUser.Password)

	token := CreateToken(loginUser, dbUser)
	if match == true {
		response.Write([]byte(`{ "message": "` + token + `"}`))
	} else {
		response.WriteHeader(http.StatusUnauthorized)
		response.Write([]byte(`{ "message": "you could not be logged in." }`))
		return
	}
}
