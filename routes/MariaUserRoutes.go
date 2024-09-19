package routes

import (
	"database/sql"

	mariadbauth "github.com/froggy-12/mooshroombase_v2/services/authentication/mariadb_auth"
	"github.com/froggy-12/mooshroombase_v2/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

func MariaUserRoutes(router fiber.Router, mariaDBClient *sql.DB) {
	validate := validator.New()
	router.Get("/get-user", func(c *fiber.Ctx) error {
		return mariadbauth.Get_User(c, mariaDBClient)
	})
	router.Put("/update-username", func(c *fiber.Ctx) error {
		return mariadbauth.UpdateUserName(c, mariaDBClient, *validate)
	})
	router.Put("/update-user-info", func(c *fiber.Ctx) error {
		return mariadbauth.UpdateUser(c, mariaDBClient)
	})
	router.Put("/update-email", func(c *fiber.Ctx) error {
		return mariadbauth.ChangeEmail(c, mariaDBClient, *validate)
	})
	router.Delete("/delete-user", func(c *fiber.Ctx) error {
		return mariadbauth.DeleteUser(c, mariaDBClient, *validate)
	})
	router.Get("/log-out", func(c *fiber.Ctx) error {
		return utils.LogOut(c)
	})
}
