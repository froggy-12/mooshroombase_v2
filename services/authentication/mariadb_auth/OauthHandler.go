package mariadbauth

import (
	"database/sql"
	"net/http"

	"github.com/froggy-12/mooshroombase_v2/configs"
	"github.com/froggy-12/mooshroombase_v2/types"
	"github.com/froggy-12/mooshroombase_v2/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func CreateOAuthUser(c *fiber.Ctx, db *sql.DB) error {
	var user types.User_Maria_Oauth_Payload
	if err := c.BodyParser(&user); err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "request body invalid"})
	}

	user_db, err := utils.FindOAuthUserFromMariaDBUsingID(user.ID, db)

	if err != nil {
		if err == sql.ErrNoRows {

			newUser := types.User_Maria_Oauth{
				ID:                user.ID,
				UserName:          user.ID,
				FirstName:         user.FirstName,
				LastName:          user.LastName,
				Email:             user.Email,
				ProfilePicture:    user.ProfilePicture,
				OAuthProvider:     user.OAuthProvider,
				Verified:          user.Verified,
				VerificationToken: uuid.New().String(),
			}
			_, err := db.Exec(`
			INSERT INTO mooshroombase.oauth_users (
				ID,
				UserName,
				FirstName,
				LastName,
				Email,
				ProfilePicture,
				Provider,
				Verified,
				VerificationToken
			) VALUES (
				?,
				?,
				?,
				?,
				?,
				?,
				?,
				?,
				?
			)
			`, newUser.ID, newUser.UserName, newUser.FirstName, newUser.LastName, newUser.Email, newUser.ProfilePicture, newUser.OAuthProvider, newUser.Verified, newUser.VerificationToken)

			if err != nil {
				return c.Status(http.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Failed to create user: " + err.Error()})
			}

			token, err := utils.GenerateJWTToken(newUser.ID, configs.Configs.HttpConfigurations.JWTTokenExpirationTime, configs.Configs.HttpConfigurations.JWTSecret)

			if err != nil {
				return c.Status(http.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Failed to generate jwt token user: " + err.Error()})
			}

			utils.SetJwtHttpCookies(c, token, configs.Configs.HttpConfigurations.CorsHeaderMaxAge)

			return c.Status(http.StatusOK).JSON(types.HttpSuccessResponse{Message: "user has been created successfully", Data: map[string]any{"userId": newUser.ID}})

		} else {
			return c.Status(http.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Something went wrong: " + err.Error()})
		}
	}

	token, err := utils.GenerateJWTToken(user_db.ID, configs.Configs.HttpConfigurations.JWTTokenExpirationTime, configs.Configs.HttpConfigurations.JWTSecret)

	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Failed to generate jwt token user: " + err.Error()})
	}

	utils.SetJwtHttpCookies(c, token, configs.Configs.HttpConfigurations.CorsHeaderMaxAge)

	return c.Status(http.StatusOK).JSON(types.HttpSuccessResponse{Message: "User has been logged in successfully"})
}

func GetOAuthUserData(c *fiber.Ctx, db *sql.DB) error {
	userId, expired, err := utils.ReadJWTToken(c.Cookies("jwtToken"), configs.Configs.HttpConfigurations.JWTSecret)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(types.ErrorResponse{Error: "User is not authorised please log in"})
	}
	if expired {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Please Log in"})
	}

	user, err := utils.FindOAuthUserFromMariaDBUsingID(userId, db)

	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Something went wrong: " + err.Error()})
	}

	return c.Status(http.StatusOK).JSON(user)
}
