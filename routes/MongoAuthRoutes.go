package routes

import (
	mongoauth "github.com/froggy-12/mooshroombase_v2/services/authentication/mongo_auth"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

var validate = validator.New()

func MongoAuthRoutes(router fiber.Router, mongoClient *mongo.Client) {
	router.Post("/create-user", func(c *fiber.Ctx) error {
		return mongoauth.CreateUserWithEmailAndPassword(c, mongoClient, *validate)
	})
	router.Post("/log-in", func(c *fiber.Ctx) error {
		return mongoauth.LogInWithEmailAndPassword(c, mongoClient, *validate)
	})
	router.Post("/send-verification-email", func(c *fiber.Ctx) error {
		return mongoauth.SendVerificationEmail(c, mongoClient)
	})
	router.Get("/verified", func(c *fiber.Ctx) error {
		return mongoauth.VerifyEmail(c, mongoClient, *validate)
	})
	router.Get("/check-email-availability", func(c *fiber.Ctx) error {
		return mongoauth.CheckIsEmailAvailable(c, mongoClient, *validate)
	})
	router.Get("/check-username-availability", func(c *fiber.Ctx) error {
		return mongoauth.CheckIsUsernameAvailable(c, mongoClient, *validate)
	})
}
