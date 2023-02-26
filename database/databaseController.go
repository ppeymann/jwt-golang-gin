package database

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DBinstance() *mongo.Client{
	
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err!= nil{
		log.Fatal(err.Error())
	}
	err = client.Connect(context.Background())
	if err!= nil{
		log.Fatal(err.Error())
	}
	err = client.Ping(context.Background(),nil)
	fmt.Println("Connect To DB")
	return client

}
var Client *mongo.Client = DBinstance()

func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection{
	var collection *mongo.Collection =(*mongo.Collection)(client.Database("jwt").Collection(collectionName))
	return collection
}