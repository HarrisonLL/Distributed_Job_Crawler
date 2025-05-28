package database

import (
	"context"
	"go_services/config"
	"log"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	MongoClient *mongo.Client
	once        sync.Once
)

func InitMongoDB() {
	once.Do(func() {
		var err error
		mongoURI, getURLErr := config.GetURL("MONGOURL")
		if getURLErr != "" {
			log.Fatal(getURLErr)
		}
		MongoClient, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
		if err != nil {
			log.Fatalf("Failed to connect to MongoDB: %v", err)
		}
	})
}
