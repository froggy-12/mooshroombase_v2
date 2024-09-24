package mongoauth

import (
	"context"
	"net/http"
	"time"

	"github.com/froggy-12/mooshroombase_v2/configs"
	"github.com/froggy-12/mooshroombase_v2/types"
	"github.com/froggy-12/mooshroombase_v2/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateOAuthUser(c *fiber.Ctx, mongoClient *mongo.Client) error {

	jwtTokenCookie := c.Cookies("jwtToken")
	if jwtTokenCookie != "" {
		userID, _, _ := utils.ReadJWTToken(jwtTokenCookie, configs.Configs.HttpConfigurations.JWTSecret)
		if userID != "" {
			return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Valid Token found please log out first then sign up"})
		}
	}

	var user types.User_Mongo_Oauth_Payload
	if err := c.BodyParser(&user); err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "request body invalid"})
	}

	coll := mongoClient.Database("mooshroombase").Collection("oauth_users")
	user_db, err := utils.FindOAuthUserFromMongoDBUsingID(user.ID, coll)

	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Something went wrong when trying to find user from database: " + err.Error()})
	}

	if user_db.ID == "" {
		newUser := types.User_Mongo_OAuth{
			ID:                user.ID,
			UserName:          user.UserName,
			FirstName:         user.FirstName,
			LastName:          user.LastName,
			Email:             user.Email,
			ProfilePicture:    user.ProfilePicture,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
			Verified:          user.Verified,
			VerificationToken: uuid.New().String(),
			RawData:           []types.RawUserData{},
			OAuthProvider:     user.OAuthProvider,
		}
		_, err = coll.InsertOne(context.Background(), newUser)

		if err != nil {
			return c.Status(http.StatusBadGateway).JSON(types.ErrorResponse{Error: "failed to create new user into the database: " + err.Error()})
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

func CheckIsEmailAvailableForOAuthUsers(c *fiber.Ctx, mongoClient *mongo.Client, validate validator.Validate) error {
	coll := mongoClient.Database("mooshroombase").Collection("oauth_users")

	email := c.Query("email")
	if email == "" {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "invalid query"})
	}
	if err := validate.Var(email, "required,email"); err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Invalid Email: " + err.Error()})
	}
	_, err := utils.FindUserFromMongoDBUsingEmail(email, coll)
	if err == nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "User already exist"})
	}
	return c.Status(http.StatusOK).JSON(types.HttpSuccessResponse{Message: "The Email is good to go"})
}

func CheckIsUsernameAvailableForOAuthUsers(c *fiber.Ctx, mongoClient *mongo.Client, validate validator.Validate) error {
	coll := mongoClient.Database("mooshroombase").Collection("oauth_users")

	username := c.Query("username")
	if username == "" {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "invalid query"})
	}
	if err := validate.Var(username, "required"); err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: err.Error()})
	}
	_, err := utils.FindUserFromMongoDBUsingUsername(username, coll)
	if err == nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "User already exist"})
	}
	return c.Status(http.StatusOK).JSON(types.HttpSuccessResponse{Message: "The username is good to go"})
}
