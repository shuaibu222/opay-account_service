package main

import (
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	newRouter := mux.NewRouter()

	AccountRoutes(newRouter)
	http.Handle("/", newRouter)

	log.Println("listening on http://localhost:9000")

	log.Fatal(http.ListenAndServe(":9000", handlers.CORS(
		handlers.AllowCredentials(),
		handlers.AllowedOrigins([]string{"http://localhost:3000"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"}),
	)(newRouter)))
}
