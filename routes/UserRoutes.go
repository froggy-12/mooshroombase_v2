package routes

import (
	mongoauth "github.com/froggy-12/mooshroombase_v2/services/authentication/mongo_auth"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

func UserRoutes(router fiber.Router, mongoClient *mongo.Client) {
	router.Get("/get-user", func(c *fiber.Ctx) error {
		return mongoauth.Get_User(c, mongoClient)
	})
	router.Put("/update-username", func(c *fiber.Ctx) error {
		return mongoauth.UpdateUserName(c, mongoClient, *validate)
	})
	router.Put("/update-user-info", func(c *fiber.Ctx) error {
		return mongoauth.UpdateUser(c, mongoClient)
	})
	router.Put("/update-email", func(c *fiber.Ctx) error {
		return mongoauth.ChangeEmail(c, mongoClient, *validator.New())
	})
	router.Put("/append-raw-data", func(c *fiber.Ctx) error {
		return mongoauth.AppendRawData(c, mongoClient)
	})
	router.Delete("/delete-user", func(c *fiber.Ctx) error {
		return mongoauth.DeleteUser(c, mongoClient, *validate)
	})
}
