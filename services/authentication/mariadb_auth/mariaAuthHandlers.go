package mariadbauth

import (
	"database/sql"
	"net/http"

	"github.com/froggy-12/mooshroombase_v2/configs"
	smtpconfigs "github.com/froggy-12/mooshroombase_v2/smtp_configs"
	"github.com/froggy-12/mooshroombase_v2/types"
	"github.com/froggy-12/mooshroombase_v2/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func CreateUserWithEmailAndPassword(c *fiber.Ctx, sqlClient *sql.DB, validate validator.Validate) error {
	jwtTokenCookie := c.Cookies("jwtToken")
	if jwtTokenCookie != "" {
		userID, _, _ := utils.ReadJWTToken(jwtTokenCookie, configs.Configs.HttpConfigurations.JWTSecret)
		if userID != "" {
			return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Valid Token found please log out first then sign up"})
		}
	}

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

	newUser := types.User_Maria{
		ID:                Id,
		UserName:          user.UserName,
		FirstName:         user.FirstName,
		LastName:          user.LastName,
		Email:             user.Email,
		Password:          string(hashedPassword),
		ProfilePicture:    "",
		Verified:          false,
		VerificationToken: verificationTokenString,
	}

	_, err = sqlClient.Exec(`
    INSERT INTO mooshroombase.users (
        ID,
        UserName,
        FirstName,
        LastName,
        Email,
        Password,
        ProfilePicture,
        CreatedAt,
        UpdatedAt,
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
        ?,
        ?,
        ?
    )
`,
		newUser.ID,
		newUser.UserName,
		newUser.FirstName,
		newUser.LastName,
		newUser.Email,
		newUser.Password,
		newUser.ProfilePicture,
		newUser.CreatedAt,
		newUser.UpdatedAt,
		newUser.Verified,
		newUser.VerificationToken,
	)

	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Failed to create user: " + err.Error()})
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

		err = smtpconfigs.SendVerificationEmail(newUser.Email, newUser.VerificationToken)
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

func LogInWithEmailAndPassword(c *fiber.Ctx, mariadbClient *sql.DB, validate validator.Validate) error {
	token := c.Cookies("jwtToken")
	if token != "" {
		userid, expired, err := utils.ReadJWTToken(token, configs.Configs.HttpConfigurations.JWTSecret)
		if err != nil || expired {
			return utils.LogInMariaDB(c, mariadbClient, validate, configs.Configs.HttpConfigurations.JWTTokenExpirationTime, configs.Configs.HttpConfigurations.JWTSecret)
		}

		_, err = utils.FindUserFromMariaDBUsingID(userid, mariadbClient)
		if err == nil {
			return c.Status(http.StatusAlreadyReported).JSON(types.HttpSuccessResponse{Message: "You are already logged in"})
		}
	}

	return utils.LogInMariaDB(c, mariadbClient, validate, configs.Configs.HttpConfigurations.JWTTokenExpirationTime, configs.Configs.HttpConfigurations.JWTSecret)

}

func SendVerificationEmail(c *fiber.Ctx, mariadbClient *sql.DB, validate validator.Validate) error {
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

	user, err := utils.FindUserFromMariaDBUsingID(body.ID, mariadbClient)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "User not found"})
		} else {
			return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Something went wrong: " + err.Error()})
		}
	}

	if user.Verified {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "User is already verified"})
	}

	newToken := uuid.New().String()

	_, err = mariadbClient.Exec(`UPDATE mooshroombase.users SET verificationToken = ? WHERE ID = ?`, newToken, user.ID)

	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(types.ErrorResponse{Error: "failed to generate and set new verification token: " + err.Error()})
	}

	err = smtpconfigs.SendVerificationEmail(user.Email, newToken)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "failed to send email to this user: " + user.Email})
	}

	return c.Status(http.StatusAccepted).JSON(types.HttpSuccessResponse{Message: "Email sent successfully"})
}

func VerifyEmail(c *fiber.Ctx, mariadbClient *sql.DB, validator validator.Validate) error {
	email := c.Query("email")
	verificationTokenString := c.Query("token")

	if email == "" || verificationTokenString == "" {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Email and token are required"})
	}

	if err := validator.Var(email, "required,email"); err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Invalid email: " + err.Error()})
	}

	user, err := utils.FindUserFromMariaDBUsingEmail(email, mariadbClient)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(http.StatusBadGateway).JSON(types.ErrorResponse{Error: "User not Found"})
		} else {
			return c.Status(http.StatusBadGateway).JSON(types.ErrorResponse{Error: "Something went wrong: " + err.Error()})
		}
	}

	if user.Verified {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "User is already verified"})
	}

	if user.VerificationToken == verificationTokenString {
		_, err := mariadbClient.Exec(`UPDATE mooshroombase.users SET Verified = true WHERE ID = ?`, user.ID)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(types.ErrorResponse{Error: "failed to update user's verification status: " + err.Error()})
		}
		return c.Status(http.StatusOK).JSON(types.HttpSuccessResponse{Message: "Email verified successfully"})
	} else {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Wrong token Provided"})
	}
}

func CheckIsEmailAvailable(c *fiber.Ctx, mariaDBClient *sql.DB, validator validator.Validate) error {
	email := c.Query("email")
	if email == "" {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "invalid query"})
	}
	if err := validator.Var(email, "required,email"); err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Invalid Email: " + err.Error()})
	}
	_, err := utils.FindUserFromMariaDBUsingEmail(email, mariaDBClient)
	if err == nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "User already exist"})
	}
	return c.Status(http.StatusOK).JSON(types.HttpSuccessResponse{Message: "The Email is good to go"})
}

func CheckIsUsernameAvailable(c *fiber.Ctx, mariaDBClient *sql.DB, validator validator.Validate) error {
	username := c.Query("username")
	if username == "" {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "invalid query"})
	}
	if err := validator.Var(username, "required"); err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: err.Error()})
	}
	_, err := utils.FindUserFromMariaDBUsingUsername(username, mariaDBClient)
	if err == nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "User already exist"})
	}
	return c.Status(http.StatusOK).JSON(types.HttpSuccessResponse{Message: "The username is good to go"})
}
