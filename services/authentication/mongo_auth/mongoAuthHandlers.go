package mongoauth

import (
	"context"
	"net/http"
	"time"

	"github.com/froggy-12/mooshroombase_v2/configs"
	smtpconfigs "github.com/froggy-12/mooshroombase_v2/smtp_configs"
	"github.com/froggy-12/mooshroombase_v2/types"
	"github.com/froggy-12/mooshroombase_v2/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

func CreateUserWithEmailAndPassword(c *fiber.Ctx, mongoClient *mongo.Client, validate validator.Validate) error {
	jwtTokenCookie := c.Cookies("jwtToken")
	if jwtTokenCookie != "" {
		userID, _, _ := utils.ReadJWTToken(jwtTokenCookie, configs.Configs.HttpConfigurations.JWTSecret)
		if userID != "" {
			return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Valid Token found please log out first then sign up"})
		}
	}

	collection := mongoClient.Database("mooshroombase").Collection("users")
	var user types.User

	if err := c.BodyParser(&user); err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Invalid Request body"})
	}

	if err := validate.Struct(&user); err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: err.Error()})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)

	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Failed to Hash the password: " + err.Error()})
	}

	Id := uuid.New().String()
	verificationTokenString := uuid.New().String()

	newUser := types.User_Mongo{
		ID:                Id,
		FirstName:         user.FirstName,
		LastName:          user.LastName,
		Email:             user.Email,
		Password:          string(hashedPassword),
		UserName:          user.UserName,
		ProfilePicture:    configs.Configs.ExtraConfigurations.DefaultProfilePictureUrl,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		Verified:          false,
		VerificationToken: verificationTokenString,
		LastLoggedIn:      types.LastTimeLoggedIn{When: time.Now()},
		RawData:           []types.RawUserData{},
	}

	_, err = collection.InsertOne(context.Background(), newUser)

	if err != nil {
		return c.Status(http.StatusBadGateway).JSON(types.ErrorResponse{Error: "failed to create new user into the database: " + err.Error()})
	}

	if configs.Configs.Authentication.SetJWTTokenAfterSignUp {
		token, err := utils.GenerateJWTToken(newUser.ID, 1, configs.Configs.HttpConfigurations.JWTSecret)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Failed to generate jwt token for user: " + newUser.ID})
		}
		utils.SetJwtHttpCookies(c, token, 1)
	}

	if configs.Configs.Authentication.SendEmailAfterSignUpWithCode {
		if !configs.Configs.Authentication.EmailVerificationAllowed {
			return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Email Verification is not configured or turned off please check again and restart the app"})
		}

		if !configs.Configs.SMTPConfigurations.SMTPEnabled {
			return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "SMTP is not configured or turned off please check again and restart the app"})
		}

		newToken := uuid.New().String()
		_, err := collection.UpdateOne(context.Background(), bson.M{"id": newUser.ID}, bson.M{"$set": bson.M{"verificationToken": newToken}})

		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "failed to update new token: " + err.Error()})
		}

		err = smtpconfigs.SendVerificationEmail(newUser.Email, newToken)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "failed to send email to this user: " + user.Email})
		}

		return c.Status(http.StatusAccepted).JSON(types.HttpSuccessResponse{Message: "User has been created successfully and sent verification email"})

	} else {
		return c.Status(http.StatusCreated).JSON(types.HttpSuccessResponse{
			Message: "User Has been created to the database hope you will verify the email first then everything",
			Data:    map[string]any{"userId": newUser.ID},
		})
	}
}

func LogInWithEmailAndPassword(c *fiber.Ctx, mongoClient *mongo.Client, validate validator.Validate) error {
	coll := mongoClient.Database("mooshroombase").Collection("users")
	token := c.Cookies("jwtToken")
	if token != "" {
		userid, expired, err := utils.ReadJWTToken(token, configs.Configs.HttpConfigurations.JWTSecret)
		if err != nil || expired {
			err = utils.LogIn(c, coll, validate, configs.Configs.HttpConfigurations.JWTTokenExpirationTime, configs.Configs.HttpConfigurations.JWTSecret)
			return err
		}

		_, err = utils.FindUserFromMongoDBUsingID(userid, coll)
		if err == nil {
			return c.Status(http.StatusAlreadyReported).JSON(types.HttpSuccessResponse{Message: "You are already logged in"})
		}

	}

	return utils.LogIn(c, coll, validate, configs.Configs.HttpConfigurations.JWTTokenExpirationTime, configs.Configs.HttpConfigurations.JWTSecret)
}

func SendVerificationEmail(c *fiber.Ctx, mongoClient *mongo.Client) error {

	if !configs.Configs.Authentication.EmailVerificationAllowed {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Email Verification is not configured or turned off please check again and restart the app"})
	}

	if !configs.Configs.SMTPConfigurations.SMTPEnabled {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "SMTP is not configured or turned off please check again and restart the app"})
	}

	var body struct {
		ID string `json:"id"`
	}

	tokenSet := c.Query("tokenSet", "false")

	if c.Cookies("jwtToken") != "" && tokenSet == "true" {
		userID, _, _ := utils.ReadJWTToken(c.Cookies("jwtToken"), configs.Configs.HttpConfigurations.JWTSecret)
		body.ID = userID
	}

	if body.ID == "" && tokenSet == "false" {
		if err := c.BodyParser(&body); err != nil {
			return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Invalid Request body"})
		}
	}

	coll := mongoClient.Database("mooshroombase").Collection("users")

	user, err := utils.FindUserFromMongoDBUsingID(body.ID, coll)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "User not found"})
	}

	newToken := uuid.New().String()
	_, err = coll.UpdateOne(context.Background(), bson.M{"id": user.ID}, bson.M{"$set": bson.M{"verificationToken": newToken}})

	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "failed to update new token: " + err.Error()})
	}

	err = smtpconfigs.SendVerificationEmail(user.Email, newToken)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "failed to send email to this user: " + user.Email})
	}

	return c.Status(http.StatusAccepted).JSON(types.HttpSuccessResponse{Message: "Email sent successfully"})
}

func VerifyEmail(c *fiber.Ctx, mongoClient *mongo.Client, validate validator.Validate) error {
	coll := mongoClient.Database("mooshroombase").Collection("users")

	email := c.Query("email")
	verificationTokenString := c.Query("token")

	if email == "" || verificationTokenString == "" {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Email and token are required"})
	}

	if err := validate.Var(email, "required,email"); err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Invalid email: " + err.Error()})
	}

	user, err := utils.FindUserFromMongoDBUsingEmail(email, coll)
	if err != nil {
		return c.Status(http.StatusBadGateway).JSON(types.ErrorResponse{Error: "User not Found"})
	}

	if user.Verified {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "User is already verified"})
	}

	if user.VerificationToken == verificationTokenString {
		_, err := coll.UpdateOne(context.Background(), bson.M{"email": email}, bson.M{"$set": bson.M{"verified": true}})
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Failed to update user verification status"})
		}
		return c.Status(http.StatusOK).JSON(types.HttpSuccessResponse{Message: "Email verified successfully"})
	} else {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Wrong token Provided"})
	}
}

func CheckIsEmailAvailable(c *fiber.Ctx, mongoClient *mongo.Client, validate validator.Validate) error {
	coll := mongoClient.Database("mooshroombase").Collection("users")

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

func CheckIsUsernameAvailable(c *fiber.Ctx, mongoClient *mongo.Client, validate validator.Validate) error {
	coll := mongoClient.Database("mooshroombase").Collection("users")

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
