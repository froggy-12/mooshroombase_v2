package db

import (
	"context"
	"database/sql"
	"log"

	"github.com/froggy-12/mooshroombase_v2/configs"
	"github.com/froggy-12/mooshroombase_v2/utils"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Init(mongoClient *mongo.Client, redisClient *redis.Client, mariaDBClient *sql.DB) {
	if configs.Configs.Authentication.Auth && configs.Configs.DatabaseConfigurations.PrimaryDB == "mongodb" {
		utils.DebugLogger("db", "detected mongodb as primary database indexing and checking some models")
		database := mongoClient.Database("mooshroombase")
		usersCollection := database.Collection("users")
		_, err := usersCollection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
			Keys:    bson.M{"email": 1},
			Options: options.Index().SetUnique(true),
		})

		if err != nil {
			log.Fatal(err)
		}

		_, err = usersCollection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
			Keys:    bson.M{"username": 1},
			Options: options.Index().SetUnique(true),
		})
		if err != nil {
			log.Fatal(err)
		}
	}
}
