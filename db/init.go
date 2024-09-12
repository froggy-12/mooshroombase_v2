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
	} else if configs.Configs.DatabaseConfigurations.PrimaryDB == "mariadb" {
		utils.DebugLogger("db", "detected mariadb as primary database running some configurations")

		_, err := mariaDBClient.Exec(`CREATE DATABASE IF NOT EXISTS mooshroombase`)

		if err != nil {
			log.Fatal(err)
		}

		_, err = mariaDBClient.Exec(`
    CREATE TABLE IF NOT EXISTS mooshroombase.users (
      ID VARCHAR(255) NOT NULL,
      UserName VARCHAR(255) NOT NULL UNIQUE,
      FirstName VARCHAR(255) NOT NULL,
      LastName VARCHAR(255) NOT NULL,
      Email VARCHAR(255) NOT NULL UNIQUE,
      Password VARCHAR(255) NOT NULL,
      ProfilePicture VARCHAR(255),
      CreatedAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
      UpdatedAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
      Verified BOOLEAN NOT NULL DEFAULT FALSE,
      VerificationToken VARCHAR(255),
      LastLoggedIn TIMESTAMP,
      PRIMARY KEY (ID)
    );
`)

		if err != nil {
			log.Fatal(err)
		}

	}
}
