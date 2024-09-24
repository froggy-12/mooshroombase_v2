package routes

import (
	"database/sql"

	mariadbauth "github.com/froggy-12/mooshroombase_v2/services/authentication/mariadb_auth"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

func MariaDBAuthRoutes(router fiber.Router, mariadbClient *sql.DB) {
	validator := validator.New()
	router.Post("/create-user", func(c *fiber.Ctx) error {
		return mariadbauth.CreateUserWithEmailAndPassword(c, mariadbClient, *validator)
	})
	router.Post("/log-in", func(c *fiber.Ctx) error {
		return mariadbauth.LogInWithEmailAndPassword(c, mariadbClient, *validator)
	})
	router.Post("/send-verification-email", func(c *fiber.Ctx) error {
		return mariadbauth.SendVerificationEmail(c, mariadbClient, *validator)
	})
	router.Get("/verified", func(c *fiber.Ctx) error {
		return mariadbauth.VerifyEmail(c, mariadbClient, *validator)
	})
	router.Get("/check-email-availability", func(c *fiber.Ctx) error {
		return mariadbauth.CheckIsEmailAvailable(c, mariadbClient, *validator)
	})
	router.Get("/check-username-availability", func(c *fiber.Ctx) error {
		return mariadbauth.CheckIsUsernameAvailable(c, mariadbClient, *validate)
	})
}
