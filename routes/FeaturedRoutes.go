package routes

import (
	"net/http"

	"github.com/froggy-12/mooshroombase_v2/configs"
	smtpconfigs "github.com/froggy-12/mooshroombase_v2/smtp_configs"
	"github.com/froggy-12/mooshroombase_v2/types"
	"github.com/froggy-12/mooshroombase_v2/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

func Pong(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(types.HttpSuccessResponse{Message: "Pong"})
}

func SendEmail(c *fiber.Ctx) error {
	var body struct {
		EmailSubject string `json:"emailSubject"`
		EmailTo      string `json:"emailTo"`
		EmailBody    string `json:"emailBody"`
	}
	validate := validator.New()

	if err := c.BodyParser(&body); err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Invalid request body"})
	}

	if err := validate.Struct(&body); err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Invalid request body: " + err.Error()})
	}

	err := smtpconfigs.SendEmailWithAnything(body.EmailSubject, body.EmailTo, body.EmailBody)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "failed to send email: " + err.Error()})
	}

	return c.Status(http.StatusAccepted).JSON(types.HttpSuccessResponse{Message: "Email has been sent successfully"})
}

func GetUserID(c *fiber.Ctx) error {
	token := c.Cookies("jwtToken")
	id, err := utils.ExtractJWTToken(token, configs.Configs.HttpConfigurations.JWTSecret)

	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Failed to pass user id: " + err.Error()})
	}

	return c.Status(http.StatusOK).JSON(types.HttpSuccessResponse{Data: map[string]any{"userID": id}})
}
