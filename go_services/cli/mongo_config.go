package cli

import (
	"context"
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

func setupTTLForAllJobcrawlerDatabases() error {
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

			// Fix string dates before setting up TTL
			fixDatetimeFormat(dbName, collection, ctx)

			ttlSeconds := int32(60 * 60 * 24 * 90) // 90 days

			// Drop existing index first
			_, err = collection.Indexes().DropOne(ctx, "ttl_crawled_datetime")
			if err != nil {
				log.Printf("Note: Could not drop existing index for %s.jobs: %v", dbName, err)
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
				log.Printf("Failed to create TTL index for %s.jobs: %v", dbName, err)
			} else {
				log.Printf("TTL index created for %s.jobs", dbName)
			}
		}
	}

	return nil
}
