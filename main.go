package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/brianvoe/sjwt"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// User defines the structure of a User
type User struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Username string             `json:"username,omitempty" bson:"username,omitempty"`
	Email    string             `json:"email,omitempty" bson:"email,omitempty"`
	Password string             `json:"password,omitempty" bson:"password,omitempty"`
}

// Token sets up a datamodel for the token
type Token struct {
	Token string `json:"token,omitempty" bson:"token,omitempty"`
}

var client *mongo.Client
var secretKey = []byte(os.Getenv("KEY"))

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

// VerifyToken serves to verify that the token is correct
func VerifyToken(token string) bool {
	hasVerified := sjwt.Verify(token, secretKey)
	claims, _ := sjwt.Parse(token)
	err := claims.Validate()
	if err != nil {
		return false
	}
	return hasVerified
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
	collection.FindOne(ctx, User{Username: signUpUser.Username}).Decode(&dbUser)
	if dbUser.Username == signUpUser.Username {
		response.Write([]byte(`{ "message": "` + "An account has already been created with that username, please choose another." + `" }`))
		return
	}
	collection.FindOne(ctx, User{Email: signUpUser.Email}).Decode(&db2User)
	if db2User.Email == signUpUser.Email {
		response.Write([]byte(`{ "message": "` + "An account has already been created with that username, please choose another." + `" }`))
		return
	}
	signUpUser.Password, _ = HashPassword(signUpUser.Password)
	result, _ := collection.InsertOne(ctx, signUpUser)
	json.NewEncoder(response).Encode(result)
}

// GetUsersEndpoint Get all Users
func GetUsersEndpoint(response http.ResponseWriter, request *http.Request) {
	var token Token
	json.NewDecoder(request.Body).Decode(&token)
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
	json.NewDecoder(request.Body).Decode(&token)
	if VerifyToken(token.Token) == false {
		response.WriteHeader(http.StatusUnauthorized)
		response.Write([]byte(`{ "message": "You need to provide a token in the body inorder to access this resource." }`))
		return
	}
	response.Header().Set("content-type", "application/json")
	params := mux.Vars(request)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var user User
	collection := client.Database("mapdb").Collection("users")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	err := collection.FindOne(ctx, User{ID: id}).Decode(&user)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(user)
}

// LoginEndpoint checks to see if a user is who they claim to be and sends them a JWT.
func LoginEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	params := mux.Vars(request)
	username := params["username"]
	var loginUser User
	var dbUser User
	json.NewDecoder(request.Body).Decode(&loginUser)
	collection := client.Database("mapdb").Collection("users")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	err := collection.FindOne(ctx, User{Username: username}).Decode(&dbUser)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	match := CheckPasswordHash(loginUser.Password, dbUser.Password)

	claims := sjwt.New()
	claims.Set("username", loginUser.Username)
	claims.Set("id", dbUser.ID.String())
	claims.SetExpiresAt(time.Now().Add(time.Hour * 24))
	claims.SetTokenID()
	token := claims.Generate(secretKey)
	if match == true {
		response.Write([]byte(`{ "message": "` + token + `"}`))
	} else {
		response.WriteHeader(http.StatusUnauthorized)
		response.Write([]byte(`{ "message": "you could not be logged in." }`))
		return
	}
}

func main() {
	fmt.Println("Starting the application...")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, _ = mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("URI")))
	router := mux.NewRouter()
	router.HandleFunc("/api/signup", CreateUserEndpoint).Methods("POST")
	router.HandleFunc("/api/user/{id}", GetUserEndpoint).Methods("GET")
	router.HandleFunc("/api/users", GetUsersEndpoint).Methods("GET")
	router.HandleFunc("/api/login", LoginEndpoint).Methods("POST")
	http.ListenAndServe(":8000", router)
}
