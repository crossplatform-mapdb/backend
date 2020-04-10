package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func main() {
	fmt.Println("Starting the application...")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, _ = mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("URI")))
	router := mux.NewRouter()

	router.HandleFunc("/api/signup", CreateUserEndpoint).Methods("POST")
	router.HandleFunc("/api/user/{id}", GetUserEndpoint).Methods("GET")
	router.HandleFunc("/api/users", GetUsersEndpoint).Methods("GET")
	router.HandleFunc("/api/login", LoginEndpoint).Methods("POST")
	router.HandleFunc("/api/place", CreatePlaceEndpoint).Methods("POST")
	router.HandleFunc("/api/places", GetPlacesEndpoint).Methods("GET")

	http.ListenAndServe(":8000", router)
}
