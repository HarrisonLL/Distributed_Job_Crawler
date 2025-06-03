package cli

import (
	"context"
	"fmt"
	"go_services/database"
	"log"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func convertStringDateToISODate(dateStr string) (time.Time, error) {
	return time.Parse("01/02/2006, 15:04:05", dateStr)
}

// Helper function to list all indexes for a collection
func listIndexes(collection *mongo.Collection, ctx context.Context, dbName string) {
	indexes, err := collection.Indexes().List(ctx)
	if err != nil {
		log.Printf("Failed to list indexes for %s.jobs: %v", dbName, err)
		return
	}

	for indexes.Next(ctx) {
		var index bson.M
		if err := indexes.Decode(&index); err != nil {
			log.Printf("Failed to decode index: %v", err)
			continue
		}
		fmt.Printf("  %v\n", index)
	}
}

// Detecting old documents with string crawled_datetime
// Update the document with the proper date object
func fixDatetimeFormat(dbName string, collection *mongo.Collection, ctx context.Context) {

	filter := bson.M{"crawled_datetime": bson.M{"$type": "string"}}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		log.Printf("Failed to query string dates in %s: %v", dbName, err)
		return
	}
	defer cursor.Close(ctx)

	count := 0
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}

		if dateStr, ok := doc["crawled_datetime"].(string); ok {
			dateObj, err := convertStringDateToISODate(dateStr)
			if err != nil {
				log.Printf("Failed to parse date '%s': %v", dateStr, err)
				continue
			}
			_, err = collection.UpdateOne(
				ctx,
				bson.M{"_id": doc["_id"]},
				bson.M{"$set": bson.M{"crawled_datetime": dateObj}},
			)
			if err == nil {
				count++
			}
		}
	}

	if count > 0 {
		log.Printf("Fixed %d documents with string dates in %s.jobs", count, dbName)
	}
}

func addUpdateTTLForAllJobcrawlerDatabases(days int) error {
	database.InitMongoDB()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client := database.MongoClient

	// List all database names
	dbs, err := client.ListDatabaseNames(ctx, bson.D{})
	if err != nil {
		return err
	}

	for _, dbName := range dbs {
		if strings.HasSuffix(dbName, "_jobcrawler") {
			collection := client.Database(dbName).Collection("jobs")
			// Fix string dates before setting up TTL, this is for backward compatible
			fixDatetimeFormat(dbName, collection, ctx)
			ttlSeconds := int32(60 * 60 * 24 * days)

			// Drop existing index first
			operation := "update"
			_, err = collection.Indexes().DropOne(ctx, "ttl_crawled_datetime")
			if err != nil {
				operation = "create"
			}

			// Create the TTL index
			indexModel := mongo.IndexModel{
				Keys: bson.D{{Key: "crawled_datetime", Value: 1}},
				Options: options.Index().
					SetExpireAfterSeconds(ttlSeconds).
					SetName("ttl_crawled_datetime"),
			}

			_, err = collection.Indexes().CreateOne(ctx, indexModel)
			if err != nil {
				log.Printf("Failed to %s TTL index for %s.jobs: %v", operation, dbName, err)
			} else {
				operationPast := operation + "d"
				log.Printf("TTL index %s to %d days for %s.jobs", operationPast, days, dbName)
			}
		}
	}

	return nil
}

func listTTLForAllJobcrawlerDatabases() error {
	database.InitMongoDB()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client := database.MongoClient
	dbs, err := client.ListDatabaseNames(ctx, bson.D{})
	if err != nil {
		return err
	}
	for _, dbName := range dbs {
		if strings.HasSuffix(dbName, "_jobcrawler") {
			collection := client.Database(dbName).Collection("jobs")
			log.Printf("Listing indexes for %s.jobs:", dbName)
			listIndexes(collection, ctx, dbName)
		}
	}
	return nil
}

func deleteTTLForAllJobcrawlerDatabases() error {
	database.InitMongoDB()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client := database.MongoClient
	dbs, err := client.ListDatabaseNames(ctx, bson.D{})
	if err != nil {
		return err
	}
	for _, dbName := range dbs {
		if strings.HasSuffix(dbName, "_jobcrawler") {
			collection := client.Database(dbName).Collection("jobs")
			log.Printf("Deleting TTL index for %s.jobs:", dbName)
			_, err = collection.Indexes().DropOne(ctx, "ttl_crawled_datetime")
			if err != nil {
				log.Printf("Failed to delete TTL index for %s.jobs: %v", dbName, err)
			} else {
				log.Printf("TTL index deleted for %s.jobs", dbName)
			}
		}
	}
	return nil
}
