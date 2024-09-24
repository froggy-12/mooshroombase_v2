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
	"golang.org/x/crypto/bcrypt"
)

func Get_User(c *fiber.Ctx, db *sql.DB) error {
	token := c.Cookies("jwtToken")
	userId, err := utils.ExtractJWTToken(token, configs.Configs.HttpConfigurations.JWTSecret)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(types.ErrorResponse{Error: "Something went Wrong: " + err.Error()})
	}

	user, err := utils.FindUserFromMariaDBUsingID(userId, db)

	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Something Went Wrong maybe user not found: " + err.Error()})
	}

	return c.Status(http.StatusOK).JSON(types.HttpSuccessResponse{
		Message: "User has been Found successfully",
		Data:    map[string]any{"user": user},
	})
}

func UpdateUser(c *fiber.Ctx, db *sql.DB) error {
	token := c.Cookies("jwtToken")
	userId, err := utils.ExtractJWTToken(token, configs.Configs.HttpConfigurations.JWTSecret)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Something went wrong: " + err.Error()})
	}

	var UpdatedUser types.UpdateMariaUser
	if err := c.BodyParser(&UpdatedUser); err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: err.Error()})
	}

	user, err := utils.FindUserFromMariaDBUsingID(userId, db)

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

	updateQuery := fmt.Sprintf("update mooshroombase.users set FirstName = '%v', LastName = '%v', ProfilePicture = '%v' WHERE ID = '%v'", UpdatedUser.FirstName, UpdatedUser.LastName, UpdatedUser.ProfilePicture, user.ID)
	_, err = db.Exec(updateQuery)

	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Failed to Update User " + err.Error()})
	}

	return c.Status(http.StatusOK).JSON(types.HttpSuccessResponse{Message: "User Has been Updated Successfully"})
}

func UpdateUserName(c *fiber.Ctx, db *sql.DB, validator validator.Validate) error {
	var body struct {
		UserName    string `json:"username" validate:"required"`
		NewUserName string `json:"newUserName" validate:"required"`
		Password    string `json:"password" validate:"required"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Invalid Request Body"})
	}

	if err := validator.Struct(&body); err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Invalid Request Body: " + err.Error()})
	}

	user, err := utils.FindUserFromMariaDBUsingUsername(body.UserName, db)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "User Not Found: " + err.Error()})
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Wrong Password or Username"})
	}

	query := fmt.Sprintf("UPDATE mooshroombase.users SET UserName = '%v' WHERE ID = '%v'", body.NewUserName, user.ID)
	_, err = db.Exec(query)

	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Failed to Update username: " + err.Error()})
	}

	return c.Status(http.StatusAccepted).JSON(types.HttpSuccessResponse{Message: "Username Has been Updated"})
}

func ChangeEmail(c *fiber.Ctx, db *sql.DB, validator validator.Validate) error {
	var body struct {
		Email    string `json:"email" validate:"required,email"`
		NewEmail string `json:"newEmail" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Invalid request body: " + err.Error()})
	}

	if err := validator.Struct(&body); err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Invalid request body: " + err.Error()})
	}

	user, err := utils.FindUserFromMariaDBUsingEmail(body.Email, db)

	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "User Not Found: " + err.Error()})
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Wrong Password or email"})
	}

	query := fmt.Sprintf("UPDATE mooshroombase.users SET Email = '%v', Verified = '%v' WHERE Email = '%v'", body.NewEmail, 0, body.Email)
	_, err = db.Exec(query)

	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Failed to update email: " + err.Error()})
	}

	return c.Status(http.StatusAccepted).JSON(types.HttpSuccessResponse{Message: "Email Has been Updated"})
}

func DeleteUser(c *fiber.Ctx, db *sql.DB, validator validator.Validate) error {
	token := c.Cookies("jwtToken")
	userId, err := utils.ExtractJWTToken(token, configs.Configs.HttpConfigurations.JWTSecret)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON("something went wrong: " + err.Error())
	}

	var body struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Invalid request body: " + err.Error()})
	}

	if err := validator.Struct(&body); err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Invalid request body: " + err.Error()})
	}

	user, err := utils.FindUserFromMariaDBUsingID(userId, db)

	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(types.ErrorResponse{Error: "User not found: " + err.Error()})
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Wrong Password or email"})
	}

	query := fmt.Sprintf("DELETE FROM mooshroombase.users WHERE Email = '%v'", body.Email)
	_, err = db.Exec(query)

	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Failed to delete user: " + err.Error()})
	}

	return c.Status(http.StatusAccepted).JSON(types.HttpSuccessResponse{Message: "User has been deleted successfully"})
}
