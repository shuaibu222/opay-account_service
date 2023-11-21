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

type Transaction struct {
	From
	To     string  `json:"to" bson:"to"`
	Amount float64 `json:"amount" bson:"amount"`
}

type From struct {
	From string `json:"from" bson:"from"`
}

type Recived struct {
	Amount float64 `json:"amount" bson:"amount"`
	To     string  `json:"to" bson:"to"`
}

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

func AddBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	acc := NewAccount{}
	money := Balance{}

	param := mux.Vars(r)

	json.NewDecoder(r.Body).Decode(&money)
	val, err := acc.AddBalanceById(param["id"], &money)
	if err != nil {
		// Handle JSON decoding error
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Invalid JSON data")
		return
	}

	data, err := json.Marshal(val)
	if err != nil {
		log.Println(err)
	}
	w.Write(data)

}

func ViewBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	acc := NewAccount{}

	param := mux.Vars(r)
	b, err := acc.GetBalanceById(param["id"])
	if err != nil {
		log.Println("Error getting balance", err)
	}

	data, err := json.Marshal(b)
	if err != nil {
		log.Println(err)
	}
	w.Write(data)
}

func SendTransactionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	p := mux.Vars(r)

	acc := NewAccount{}
	transaction := Transaction{}

	err := json.NewDecoder(r.Body).Decode(&transaction)
	if err != nil {
		log.Println(err)
	}

	accountFounded, err := acc.GetAccountById(p["id"])
	if err != nil {
		log.Println("finding account failed", err)
	}

	// update users account no. from the id
	transaction.From.From = accountFounded.AccountNo

	// send it to rabbitmq queue
	err = SendTransaction("send", transaction)
	if err != nil {
		log.Println(err)
	} else {
		accountFounded.AccountBalance.Balance -= transaction.Amount
		u, err := collection.UpdateOne(context.TODO(), bson.M{"_id": p["id"]}, bson.M{"$set": bson.M{"balance": accountFounded.AccountBalance.Balance}})
		if err != nil {
			log.Println(err)
		}

		json.NewEncoder(w).Encode(u)
	}

	json.NewEncoder(w).Encode("Processing transaction...")
}

func GetByAccountNo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	rcv := Recived{}
	acc := NewAccount{}

	err := json.NewDecoder(r.Body).Decode(&rcv)
	if err != nil {
		log.Println(err)
	}

	accountFounded, err := acc.GetUserByAccount(rcv.To)
	if err != nil {
		log.Println("finding account failed", err)
	}

	accountFounded.AccountBalance.Balance += rcv.Amount
	u, err := collection.UpdateOne(context.TODO(), bson.M{"_id": accountFounded.ID}, bson.M{"$set": bson.M{"balance": accountFounded.AccountBalance.Balance}})
	if err != nil {
		log.Println(err)
	}

	json.NewEncoder(w).Encode(u)

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
