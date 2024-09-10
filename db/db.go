package db

import (
	"context"
	"database/sql"
	"log"

	"github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectToMongoDB(mongoDBURI string) *mongo.Client {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoDBURI))
	if err != nil {
		log.Fatal("Failed to connect with MongoDB 🍃: " + err.Error())
	}
	return client
}

func ConnectToRedisDB(addr, password string) *redis.Client {
	options := &redis.Options{
		Addr:     addr,
		Password: password,
	}

	client := redis.NewClient(options)
	_, err := client.Ping(context.Background()).Result()

	if err != nil {
		log.Fatal("Failed to connect with redis 🔴: " + err.Error())
	}

	return client
}

func ConnectToMariaDB(password, address string) *sql.DB {
	cfg := mysql.Config{
		User:                 "root",
		Passwd:               password,
		Addr:                 address,
		Net:                  "tcp",
		AllowNativePasswords: true,
		ParseTime:            true,
	}

	db, err := sql.Open("mysql", cfg.FormatDSN())

	if err != nil {
		log.Fatal("Failed to create mariadb client instance: " + err.Error())
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Error Connecting to MariaDB 🐬: ", err.Error())
	}

	return db
}
