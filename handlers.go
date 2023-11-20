package main

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/bson"
)

var client *amqp.Connection

func connect() (*amqp.Connection, error) {
	var counts int64
	var backOff = 1 * time.Second
	var connection *amqp.Connection

	// don't continue until rabbit is ready
	for {
		client, err := amqp.Dial("amqp://guest:guest@rabbitmq")
		if err != nil {
			log.Println("RabbitMQ not yet ready...")
			counts++
		} else {
			log.Println("Connected to RabbitMQ!")
			connection = client
			break
		}

		if counts > 5 {
			log.Println(err)
			return nil, err
		}

		backOff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		log.Println("backing off...")
		time.Sleep(backOff)
		continue
	}

	return connection, nil
}

func CreateAccount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	acc := &NewAccount{}

	err := json.NewDecoder(r.Body).Decode(&acc)
	if err != nil {
		// Handle JSON decoding error
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Invalid JSON data")
		return
	}

	newAccount, err := acc.CreateNewAccount()
	if err != nil {
		log.Println("create account failed", err)
	}

	err = collection.FindOne(context.Background(), bson.M{"_id": newAccount.InsertedID}).Decode(&acc)
	if err != nil {
		log.Println("failed to find account", err)
	}

	client, err = connect()
	if err != nil {
		log.Println(err)
	}

	defer client.Close()

	SendToRabbitmq("accountCreated", acc)

	data, err := json.Marshal(acc)
	if err != nil {
		log.Println(err)
	}
	w.Write(data)
}

func GetAccount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	p := mux.Vars(r)

	acc := NewAccount{}

	accountFounded, err := acc.GetAccountById(p["id"])
	if err != nil {
		log.Println("finding account failed", err)
	}

	data, err := json.Marshal(accountFounded)
	if err != nil {
		log.Println(err)
	}
	w.Write(data)
}

func UpdateAccount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	p := mux.Vars(r)

	acc := NewAccount{}

	founded, err := acc.GetAccountById(p["id"])
	if err != nil {
		log.Println(err)
	}

	json.NewDecoder(r.Body).Decode(&acc)

	if acc.FullName != "" {
		founded.FullName = acc.FullName
	}
	if acc.AccountNo != "" {
		founded.AccountNo = acc.AccountNo
	}
	if acc.Email != "" {
		founded.Email = acc.Email
	}
	if acc.Password != "" {
		founded.Password = acc.Password
	}
	if acc.BVN != "" {
		founded.BVN = acc.BVN
	}

	updatedAccount, err := acc.UpdateAccountById(p["id"], acc)
	if err != nil {
		log.Println("updating account failed", err)
	}

	client, err = connect()
	if err != nil {
		log.Println(err)
	}

	defer client.Close()

	SendToRabbitmq("accountUpdated", updatedAccount)

	data, err := json.Marshal(updatedAccount)
	if err != nil {
		log.Println(err)
	}
	w.Write(data)
}

func DeleteAccount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	p := mux.Vars(r)

	acc := NewAccount{}

	deleteCount, err := acc.DeleteAccount(p["id"])
	if err != nil {
		log.Println("updating account failed", err)
	}

	client, err = connect()
	if err != nil {
		log.Println(err)
	}

	defer client.Close()

	SendToRabbitmq("accountDeleted", deleteCount)

	data, err := json.Marshal(deleteCount)
	if err != nil {
		log.Println(err)
	}

	w.Write(data)
}
