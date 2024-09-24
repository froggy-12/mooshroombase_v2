package middlewares

import (
	"net/http"
	"strings"
	"time"

	"github.com/froggy-12/mooshroombase_v2/configs"
	"github.com/froggy-12/mooshroombase_v2/types"
	"github.com/froggy-12/mooshroombase_v2/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

var CorsMiddleWare = cors.New(cors.Config{
	AllowOrigins: strings.Join(configs.Configs.Applications.AllowedCorsOrigins, ", "),
	AllowHeaders: "*",
	AllowMethods: strings.Join([]string{"GET", "POST", "PUT", "DELETE", "PATCH"}, ", "),
	MaxAge:       time.Now().Hour() * 24 * configs.Configs.HttpConfigurations.CorsHeaderMaxAge,
})

func CheckAndRefreshJWTTokenMiddleware(c *fiber.Ctx) error {
	userId, expired, err := utils.ReadJWTToken(c.Cookies("jwtToken"), configs.Configs.HttpConfigurations.JWTSecret)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(types.ErrorResponse{Error: "User is not authorised please log in"})
	}
	if expired {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Please Log in"})
	}
	// Pass the token instead of the user ID
	newToken, err := utils.RefreshJWTToken(c.Cookies("jwtToken"), configs.Configs.HttpConfigurations.JWTSecret, configs.Configs.HttpConfigurations.JWTTokenExpirationTime)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(types.ErrorResponse{
			Error: "Failed to refresh JWT token",
		})
	}

	utils.SetJwtHttpCookies(c, newToken, configs.Configs.HttpConfigurations.JWTTokenExpirationTime)

	c.Locals("userId", userId)

	return c.Next()
}
