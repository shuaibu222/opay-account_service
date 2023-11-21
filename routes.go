package main

import (
	"github.com/gorilla/mux"
)

func AccountRoutes(router *mux.Router) {
	router.HandleFunc("/accounts", CreateAccount).Methods("POST")
	router.HandleFunc("/accounts/{id}", GetAccount).Methods("GET")
	router.HandleFunc("/accounts/{id}", UpdateAccount).Methods("PUT")
	router.HandleFunc("/accounts/{id}", DeleteAccount).Methods("DELETE")
	router.HandleFunc("/balance/add/{id}", AddBalance).Methods("PUT")
	router.HandleFunc("/balance/view/{id}", ViewBalance).Methods("GET")
	router.HandleFunc("/send/{id}", SendTransactionHandler).Methods("GET")
}
