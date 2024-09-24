package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/froggy-12/mooshroombase_v2/api"
	"github.com/froggy-12/mooshroombase_v2/configs"
	"github.com/froggy-12/mooshroombase_v2/db"
	"github.com/froggy-12/mooshroombase_v2/docker"
	"github.com/froggy-12/mooshroombase_v2/utils"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	mongoClient   *mongo.Client
	mariaDBClient *sql.DB
	redisClient   *redis.Client
)

func main() {

	fmt.Println(`                                                                                          
                                     .:=+********+=-:                                     
                                .=*%@@@@@%%%%%%%@@@@@@%*=:                                
                            .=*@@@%##************####%%@@@@#=.                            
                         .=%@@%#********************######%@@@%=.                         
                       :*@@%@@************************######@@%@@#:                       
                     :#@@*: @@#************************####%@@:-*@@#:                     
                   .*@@*.   @@%**************************##%@@...:*@@#.                   
                  =@@%:     @@#****************************%@@:....:#@@-                  
                 *@@=      *@@***********##%%%%%##**********@@*......+@@*                 
                #@@:      +@@#*******#%@@@%#**#%@@@@%*******#@@+:.....-@@%.               
               %@%      -%@@#******#@@%=:         -*@@%*******@@%+:....:@@%.              
              #@@--=+*%@@%#*******#@@-              .%@%+******%@@@%#*+==@@%              
             +@@@@@@%%#***********@@-                :@@*******###%%%@@@@@@@#             
             @@%******************@@-                =@@********##########%@@:            
            -@@*******************#@@=              =@@#*******############@@=            
            -@@********************#@@@+-.      .-*@@@*********############@@=            
            :@@#*********************#%@@@@@%@@@@@@%**********############%@@-            
             +@@#*************************#####*************#############%@@*             
              -@@@#***************************************#############%@@@=              
                -#@@@%%##******************************##########%%%@@@@#=                
                   :=+*%@@@@@@@@@@@@@@%%%%%%%%%%%%@@@@@@@@@@@@@@@@%#+=:                   
                           .::::-=++++++++++++++++++++++--::::.                           
                                 ==..................:=-                                  
                                :@@-.................*@%                                  
                                :@@-.................*@%                                  
                                :@@-.................*@%                                  
                                :@@-.................*@%                                  
                                :@@-.................*@%                                  
                                :@@#+++++++++++++++++%@%                                  
                                 :+*******************=                                   
                                                                                        `)

	fmt.Println(`______ Mooshroombase ______
|       Version 2       |
|  Copyright (c) by froggy-12  |
|  [github.com/froggy-12](https://github.com/froggy-12)  |
______ ______ ______`)
	fmt.Println()
	fmt.Println()

	fmt.Println("Thanks for giving a try to mooshroombase <3 app is starting in 5 seconds")
	time.Sleep(5 * time.Second)

	// initializing configurations
	fmt.Println("initializing configurations 📄")
	configs.Configs = configs.InitConfigs()
	fmt.Println("checking configurations 📃")
	configs.CheckIfFieldsAreEmpty(configs.Configs)
	fmt.Println("Configurations Done Starting the app.....😊")

	utils.DebugLogging = configs.Configs.ExtraConfigurations.DebugLogging
	if !configs.Configs.ExtraConfigurations.DebugLogging {
		fmt.Println("Starting....")
	}

	// initializing docker
	utils.DebugLogger("main", "Configurations Done Starting the app.....😊")
	docker.Init()
	utils.DebugLogger("main", "Docker initialization completed")

	// initializing database connections
	utils.DebugLogger("main", "Starting Database initialization")
	for _, database := range configs.Configs.DatabaseConfigurations.RunningDatabases {
		switch database {
		case "mongodb":
			utils.DebugLogger("main", "connecting to MongoDB 🍃")
			var mongoURI string = fmt.Sprintf("mongodb://root:%v@localhost:%v", configs.Configs.DatabaseConfigurations.MongoDBRootPassword, configs.Configs.DatabaseConfigurations.MongoDBServerPort)
			if configs.Configs.Authentication.RealTimeUserData {
				mongoURI = fmt.Sprintf("mongodb://root:%v@127.0.0.1:%v/?directConnection=true&serverSelectionTimeoutMS=2000", configs.Configs.DatabaseConfigurations.MongoDBRootPassword, configs.Configs.DatabaseConfigurations.MongoDBServerPort)
			}
			mongoClient = db.ConnectToMongoDB(mongoURI)
			utils.DebugLogger("main", "Connected to MongoDB 🍃🍃")
		case "redis":
			utils.DebugLogger("main", "connecting to Redis 🔴")
			redisURI := fmt.Sprintf("localhost:%v", configs.Configs.DatabaseConfigurations.RedisDBServerPort)
			redisClient = db.ConnectToRedisDB(redisURI, configs.Configs.DatabaseConfigurations.RedisDBRootPassword)
			utils.DebugLogger("main", "Connected to Redis 🔴🔴")
		case "mariadb":
			utils.DebugLogger("main", "connecting to MariaDB 🐬")
			mariaDBURI := fmt.Sprintf("localhost:%v", configs.Configs.DatabaseConfigurations.MariaDBServerPort)
			mariaDBClient = db.ConnectToMariaDB(configs.Configs.DatabaseConfigurations.MariaDBRootPassword, mariaDBURI)
			utils.DebugLogger("main", "Connected to MariaDB 🐬🐬")
		}
	}
	utils.DebugLogger("main", "Database Connections are successfull 😊😊")

	// initializing database configs
	db.Init(mongoClient, redisClient, mariaDBClient)

	// Starting The API Server
	utils.DebugLogger("main", "Starting The API Server 🎉🎉🎉🍾💥")
	server := api.NewAPIServer(configs.Configs.Applications.BackEndPort, mongoClient, redisClient, mariaDBClient)
	err := server.Start()
	if err != nil {
		log.Fatal("Failed to start API Server: " + err.Error())
	}

}
