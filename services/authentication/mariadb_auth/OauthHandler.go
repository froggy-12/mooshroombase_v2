package mariadbauth

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/froggy-12/mooshroombase_v2/configs"
	"github.com/froggy-12/mooshroombase_v2/types"
	"github.com/froggy-12/mooshroombase_v2/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func CreateOAuthUser(c *fiber.Ctx, db *sql.DB) error {
	jwtTokenCookie := c.Cookies("jwtToken")
	if jwtTokenCookie != "" {
		userID, _, _ := utils.ReadJWTToken(jwtTokenCookie, configs.Configs.HttpConfigurations.JWTSecret)
		if userID != "" {
			return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Valid Token found please log out first then sign up"})
		}
	}
	var user types.User_Maria_Oauth_Payload
	if err := c.BodyParser(&user); err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "request body invalid"})
	}

	user_db, err := utils.FindOAuthUserFromMariaDBUsingID(user.ID, db)

	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Something went wrong: " + err.Error()})
	}

	if user_db.ID == "" {

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

func UpdateOAuthUserData(c *fiber.Ctx, db *sql.DB) error {
	token := c.Cookies("jwtToken")
	userId, err := utils.ExtractJWTToken(token, configs.Configs.HttpConfigurations.JWTSecret)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Something went wrong: " + err.Error()})
	}

	var UpdatedUser types.UpdateMariaUser
	if err := c.BodyParser(&UpdatedUser); err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: err.Error()})
	}

	user, err := utils.FindOAuthUserFromMariaDBUsingID(userId, db)

	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "User Not Found: " + err.Error()})
	}

	if UpdatedUser.FirstName == "" {
		UpdatedUser.FirstName = user.FirstName
	}
	if UpdatedUser.LastName == "" {
		UpdatedUser.LastName = user.LastName
	}
	if UpdatedUser.ProfilePicture == "" {
		UpdatedUser.ProfilePicture = user.ProfilePicture
	}

	updateQuery := fmt.Sprintf("update mooshroombase.oauth_users set FirstName = '%v', LastName = '%v', ProfilePicture = '%v' WHERE ID = '%v'", UpdatedUser.FirstName, UpdatedUser.LastName, UpdatedUser.ProfilePicture, user.ID)

	_, err = db.Exec(updateQuery)

	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Failed to Update User " + err.Error()})
	}

	return c.Status(http.StatusOK).JSON(types.HttpSuccessResponse{Message: "User Has been Updated Successfully"})
}

func CheckIsAvailableForOAuthUser(c *fiber.Ctx, db *sql.DB, validator validator.Validate) error {
	username := c.Query("username")
	if username == "" {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "invalid query"})
	}
	if err := validator.Var(username, "required"); err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: err.Error()})
	}
	_, err := utils.FindOAuthUserFromMariaDBUsingUserName(username, db)
	if err == nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "User already exist"})
	}
	return c.Status(http.StatusOK).JSON(types.HttpSuccessResponse{Message: "The username is good to go"})
}

func CheckIsEmailAvailableForOAuthUser(c *fiber.Ctx, db *sql.DB, validator validator.Validate) error {
	email := c.Query("email")
	if email == "" {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "invalid query"})
	}
	if err := validator.Var(email, "required,email"); err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Invalid Email: " + err.Error()})
	}
	_, err := utils.FindOAuthUserFromMariaDBUsingEmail(email, db)
	if err == nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "User already exist"})
	}
	return c.Status(http.StatusOK).JSON(types.HttpSuccessResponse{Message: "The Email is good to go"})
}
