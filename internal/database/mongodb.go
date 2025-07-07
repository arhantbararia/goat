package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoService struct {
	client *mongo.Client
}

var (
	mongoURI      = os.Getenv("BLUEPRINT_MONGO_URI")
	mongoInstance *mongoService
)

func NewMongo() Service {
	if mongoInstance != nil {
		return mongoInstance
	}
	if mongoURI == "" {
		log.Fatal("BLUEPRINT_MONGO_URI environment variable not set")
	}
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := client.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	mongoInstance = &mongoService{
		client: client,
	}
	return mongoInstance
}

func (s *mongoService) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)
	err := s.client.Ping(ctx, nil)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf(fmt.Sprintf("db down: %v", err))
		return stats
	}

	stats["status"] = "up"
	stats["message"] = "It's healthy"
	// MongoDB Go driver does not expose connection pool stats directly.
	return stats
}

func (s *mongoService) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	log.Printf("Disconnected from MongoDB: %s", mongoURI)
	return s.client.Disconnect(ctx)
}

func (s *mongoService) Create(ctx context.Context, table string, data map[string]interface{}) (interface{}, error) {
	collection := s.client.Database("").Collection(table)
	res, err := collection.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}
	return res.InsertedID, nil
}

func (s *mongoService) Read(ctx context.Context, table string, filter map[string]interface{}) ([]map[string]interface{}, error) {
	collection := s.client.Database("").Collection(table)
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []map[string]interface{}
	for cursor.Next(ctx) {
		var doc map[string]interface{}
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		results = append(results, doc)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func (s *mongoService) Update(ctx context.Context, table string, filter map[string]interface{}, update map[string]interface{}) (int64, error) {
	collection := s.client.Database("").Collection(table)
	res, err := collection.UpdateMany(ctx, filter, map[string]interface{}{"$set": update})
	if err != nil {
		return 0, err
	}
	return res.ModifiedCount, nil
}

func (s *mongoService) Delete(ctx context.Context, table string, filter map[string]interface{}) (int64, error) {
	collection := s.client.Database("").Collection(table)
	res, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}
	return res.DeletedCount, nil
}

func (s *mongoService) Exec(ctx context.Context, query string, args ...interface{}) (interface{}, error) {
	// MongoDB does not support raw SQL queries like SQL databases.
	// This method can be used for administrative commands or aggregation pipelines.
	return nil, fmt.Errorf("Exec method is not supported in MongoDB")
}
func (s *mongoService) Query(ctx context.Context, query string, args ...interface{}) ([]map[string]interface{}, error) {
	// MongoDB does not support raw SQL queries like SQL databases.
	// This method can be used for administrative commands or aggregation pipelines.
	return nil, fmt.Errorf("Query method is not supported in MongoDB")
}
