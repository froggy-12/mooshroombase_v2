package routes

import (
	"github.com/gofiber/fiber/v2"
)

func FreeRoutes(router fiber.Router) {
	router.Get("/ping", Pong)
}
