package api

import (
	"database/sql"
	"log"

	"github.com/froggy-12/mooshroombase_v2/configs"
	"github.com/froggy-12/mooshroombase_v2/middlewares"
	"github.com/froggy-12/mooshroombase_v2/routes"
	mongoauth "github.com/froggy-12/mooshroombase_v2/services/authentication/mongo_auth"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

type Server struct {
	addr          string
	mongoClient   *mongo.Client
	redisClient   *redis.Client
	mariaDBClient *sql.DB
}

func NewAPIServer(addr string, mongClient *mongo.Client, redisClient *redis.Client, mariaDBClient *sql.DB) *Server {
	return &Server{
		addr:          addr,
		mongoClient:   mongClient,
		redisClient:   redisClient,
		mariaDBClient: mariaDBClient,
	}
}

func (s *Server) Start() error {
	app := fiber.New(fiber.Config{
		BodyLimit:       configs.Configs.ExtraConfigurations.BodySizeLimit,
		ServerHeader:    "HTTPS",
		Concurrency:     256 * 1024,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	})

	if configs.Configs.ExtraConfigurations.DebugLogging {
		app.Use(logger.New())
	}

	if configs.Configs.ExtraConfigurations.RealTimeMainSwitch {
		app.Use("/ws", func(c *fiber.Ctx) error {
			if websocket.IsWebSocketUpgrade(c) {
				c.Locals("allowed", true)
				return c.Next()
			}
			return fiber.ErrUpgradeRequired
		})
		if configs.Configs.DatabaseConfigurations.PrimaryDB == "mongodb" {
			if configs.Configs.Authentication.RealTimeUserData {
				app.Use("/ws/api/user/get-user", websocket.New(func(c *websocket.Conn) {
					mongoauth.GetRealTimeUserData(c, s.mongoClient)
				}))
			}
		}
	}

	// groups
	freeRouter := app.Group("/api", middlewares.CorsMiddleWare)
	fileUploadingrouter := app.Group("/api", middlewares.CorsMiddleWare)

	if configs.Configs.SMTPConfigurations.SMTPAllowedForEveryone {
		app.Post("/api/email/send-email", routes.SendEmail)
	} else {
		router := app.Group("/api/email", middlewares.CheckAndRefreshJWTTokenMiddleware)
		router.Post("/send-email", routes.SendEmail)
	}

	// routes
	routes.FreeRoutes(freeRouter)
	if configs.Configs.Features.FileUplaod {
		routes.FileUploadingRoutes(fileUploadingrouter)
	}

	// mongodb auth routes
	if configs.Configs.Authentication.Auth {
		if configs.Configs.DatabaseConfigurations.PrimaryDB == "mongodb" {
			found := false
			for _, db := range configs.Configs.DatabaseConfigurations.RunningDatabases {
				if db == "mongodb" {
					found = true
					break
				}
			}
			if found {
				// mongo auth routes
				authRouter := app.Group("/api/auth")
				userRouter := app.Group("/api/data", middlewares.CheckAndRefreshJWTTokenMiddleware)
				routes.MongoAuthRoutes(authRouter, s.mongoClient)
				routes.UserRoutes(userRouter, s.mongoClient)
				app.Get("/api/auth/user-id", middlewares.CheckAndRefreshJWTTokenMiddleware, routes.GetUserID)
			} else {
				log.Fatal("primary database set to mongodb but its not even running")
			}
		}
	}

	return app.Listen(s.addr)
}
