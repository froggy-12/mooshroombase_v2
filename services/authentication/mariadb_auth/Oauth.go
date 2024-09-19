package mariadbauth

import (
	"database/sql"
	"net/http"

	"github.com/froggy-12/mooshroombase_v2/configs"
	"github.com/froggy-12/mooshroombase_v2/middlewares"
	"github.com/froggy-12/mooshroombase_v2/types"
	"github.com/froggy-12/mooshroombase_v2/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
	"github.com/shareed2k/goth_fiber"
)

func OAuth(router fiber.Router, db *sql.DB) {
	goth.UseProviders(
		github.New(configs.Configs.Authentication.GithubOAuthAppID, configs.Configs.Authentication.GithubOAuthAppSecret, configs.Configs.Applications.BackEndURlWithDomain+"/api/auth/oauth/callback/github"),
		google.New(configs.Configs.Authentication.GoogleOAuthAppID, configs.Configs.Authentication.GoogleOAuthAppSecret, configs.Configs.Applications.BackEndURlWithDomain+"/api/auth/oauth/callback/google"),
	)

	router.Get("/login/:provider", goth_fiber.BeginAuthHandler)
	router.Get("/callback/:provider", func(c *fiber.Ctx) error {
		user, err := goth_fiber.CompleteUserAuth(c)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(types.ErrorResponse{Error: "something went wrong: " + err.Error()})
		}
		return c.Status(http.StatusOK).JSON(types.HttpSuccessResponse{Message: "user logged in please make sure to check if its new user or not add a onboarding system to your client app", Data: map[string]any{"user": user}})
	})

	router.Get("/logout", func(c *fiber.Ctx) error {
		if err := goth_fiber.Logout(c); err != nil {
			return c.Status(http.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Failed to log out: " + err.Error()})
		}
		return c.SendString("User has been logged out")
	})

	router.Post("/create-oauth-user", func(c *fiber.Ctx) error {
		return CreateOAuthUser(c, db)
	})

	router.Get("/log-out", utils.LogOut)
	router.Get("/get-user", middlewares.CheckAndRefreshJWTTokenMiddleware, func(c *fiber.Ctx) error {
		return GetOAuthUserData(c, db)
	})
}