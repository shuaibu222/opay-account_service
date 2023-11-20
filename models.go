package main

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var collection *mongo.Collection

// const mongoURL = "mongodb://host.minikube.internal:27017"
const mongoURL = "mongodb://mongo:27017"

// connect with MongoDB
func init() {
	credential := options.Credential{
		Username: "opay",
		Password: "opay_password",
	}
	clientOpts := options.Client().ApplyURI(mongoURL).SetAuth(credential)
	client, err := mongo.Connect(context.TODO(), clientOpts)
	if err != nil {
		log.Println("Error connecting to MongoDB")
		return
	}

	collection = client.Database("account").Collection("account")

	// collection instance
	log.Println("Collections instance is ready")
}

type NewAccount struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	FullName  string             `json:"fullname" bson:"fullname"`
	AccountNo string             `json:"account" bson:"account"`
	Email     string             `bson:"email" json:"email"`
	Password  string             `bson:"password" json:"password"`
	BVN       string             `bson:"bvn" json:"bvn"`
}

func (u *NewAccount) CreateNewAccount() (*mongo.InsertOneResult, error) {
	account, err := collection.InsertOne(context.Background(), u)
	if err != nil {
		log.Println("Error while creating account:", err)
	}

	return account, nil
}

func (u *NewAccount) GetAccountById(id string) (*NewAccount, error) {
	var account *NewAccount
	Id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Println("Failed to convert id", err)
		return nil, err
	}

	filter := bson.M{"_id": Id}
	err = collection.FindOne(context.Background(), filter).Decode(&account)
	if err != nil {
		log.Println("Failed to get account", err)
		return nil, err
	}

	return account, nil
}

// for authentication purposes only
func (u *NewAccount) GetUserByUsername(username string) (NewAccount, error) {
	var user NewAccount
	filter := bson.M{"username": username}
	err := collection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		log.Println("Failed to decode account: ", err)
	}

	return user, nil
}

func (u *NewAccount) UpdateAccountById(id string, accountToUpdate NewAccount) (*mongo.UpdateResult, error) {
	Id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Println("Failed to convert id", err)
		return nil, err
	}

	updatedCount, err := collection.UpdateOne(context.Background(), bson.M{"_id": Id}, bson.M{"$set": accountToUpdate})
	if err != nil {
		log.Println("Failed to update the account: ", err)
		return nil, err
	}
	return updatedCount, nil
}

func (u *NewAccount) DeleteAccount(id string) (*mongo.DeleteResult, error) {
	Id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Println("Failed to convert id", err)
		return nil, err
	}

	deleteCount, err := collection.DeleteOne(context.Background(), bson.M{"_id": Id})
	if err != nil {
		log.Println("Failed to delete account", err)
		return nil, err
	}

	return deleteCount, nil
}
