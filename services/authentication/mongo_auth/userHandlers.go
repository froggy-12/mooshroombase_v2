package mongoauth

import (
	"context"
	"net/http"
	"time"

	"github.com/froggy-12/mooshroombase_v2/configs"
	"github.com/froggy-12/mooshroombase_v2/types"
	"github.com/froggy-12/mooshroombase_v2/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

func Get_User(c *fiber.Ctx, mongoClient *mongo.Client) error {
	token := c.Cookies("jwtToken")
	userId, err := utils.ExtractJWTToken(token, configs.Configs.HttpConfigurations.JWTSecret)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(types.ErrorResponse{Error: "Something went Wrong: " + err.Error()})
	}

	coll := mongoClient.Database("mooshroombase").Collection("users")
	user, err := utils.FindUserFromMongoDBUsingID(userId, coll)

	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Something Went Wrong maybe user not found: " + err.Error()})
	}

	return c.Status(http.StatusOK).JSON(types.HttpSuccessResponse{
		Message: "User has been Found successfully",
		Data:    map[string]any{"user": user},
	})
}

func UpdateUser(c *fiber.Ctx, mongoClient *mongo.Client) error {
	token := c.Cookies("jwtToken")
	userId, err := utils.ExtractJWTToken(token, configs.Configs.HttpConfigurations.JWTSecret)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Something went wrong: " + err.Error()})
	}

	var UpdatedUser types.UpdateMongoUser
	if err := c.BodyParser(&UpdatedUser); err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: err.Error()})
	}

	coll := mongoClient.Database("mooshroombase").Collection("users")
	user, err := utils.FindUserFromMongoDBUsingID(userId, coll)

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
	if len(UpdatedUser.RawData) == 0 {
		UpdatedUser.RawData = user.RawData
	}

	_, err = coll.UpdateOne(context.Background(), bson.M{"id": user.ID}, bson.M{"$set": bson.M{"firstName": UpdatedUser.FirstName, "lastName": UpdatedUser.LastName, "profilePicture": UpdatedUser.ProfilePicture, "rawData": UpdatedUser.RawData, "updatedAt": time.Now()}})

	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Failed to Update User " + err.Error()})
	}

	return c.Status(http.StatusOK).JSON(types.HttpSuccessResponse{Message: "User Has been Updated Successfully"})
}

func UpdateUserName(c *fiber.Ctx, mongoClient *mongo.Client, validator validator.Validate) error {
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

	coll := mongoClient.Database("mooshroombase").Collection("users")
	user, err := utils.FindUserFromMongoDBUsingUsername(body.UserName, coll)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "User Not Found: " + err.Error()})
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Wrong Password or Username"})
	}

	_, err = coll.UpdateOne(context.Background(), bson.M{"id": user.ID}, bson.M{"$set": bson.M{"username": body.NewUserName, "updatedAt": time.Now()}})
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Something went wrong failed to update username: " + err.Error()})
	}

	return c.Status(http.StatusAccepted).JSON(types.HttpSuccessResponse{Message: "Username Has been Updated"})
}

func AppendRawData(c *fiber.Ctx, mongoClient *mongo.Client) error {
	token := c.Cookies("jwtToken")
	userId, err := utils.ExtractJWTToken(token, configs.Configs.HttpConfigurations.JWTSecret)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(types.ErrorResponse{Error: "Something went Wrong: " + err.Error()})
	}

	coll := mongoClient.Database("mooshroombase").Collection("users")

	var body types.RawUserData
	var requestBody map[string]any
	if err := c.BodyParser(&requestBody); err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: err.Error()})
	}

	body.Data = requestBody

	_, err = coll.UpdateOne(context.Background(), bson.M{"id": userId}, bson.M{"$push": bson.M{"rawData": body}, "$set": bson.M{"updatedAt": time.Now()}})

	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Failed to append raw data: " + err.Error()})
	}

	return c.Status(http.StatusOK).JSON(types.HttpSuccessResponse{Message: "Raw data appended successfully"})
}

func ChangeEmail(c *fiber.Ctx, mongoClient *mongo.Client, validator validator.Validate) error {
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
	coll := mongoClient.Database("mooshroombase").Collection("users")

	user, err := utils.FindUserFromMongoDBUsingEmail(body.Email, coll)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "User Not Found: " + err.Error()})
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Wrong Password or email"})
	}
	_, err = coll.UpdateOne(context.Background(), bson.M{"id": user.ID}, bson.M{"$set": bson.M{"email": body.NewEmail, "verified": false, "updatedAt": time.Now()}})
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Something went wrong failed to update email: " + err.Error()})
	}

	return c.Status(http.StatusAccepted).JSON(types.HttpSuccessResponse{Message: "Email Has been Updated"})
}

func DeleteUser(c *fiber.Ctx, mongoClient *mongo.Client, validator validator.Validate) error {
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
	coll := mongoClient.Database("mooshroombase").Collection("users")
	user, err := utils.FindUserFromMongoDBUsingID(userId, coll)

	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(types.ErrorResponse{Error: "User not found: " + err.Error()})
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Wrong Password or email"})
	}

	_, err = coll.DeleteOne(context.Background(), bson.M{"email": user.Email})

	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Failed to delete user: " + err.Error()})
	}

	return c.Status(http.StatusAccepted).JSON(types.HttpSuccessResponse{Message: "User has been deleted successfully"})
}

func GetRealTimeUserData(c *websocket.Conn, mongoClient *mongo.Client) {
	userId := c.Query("user_id")

	if userId == "" {
		c.WriteJSON(types.ErrorResponse{Error: "User ID is required"})
		c.Close()
		return
	}
	coll := mongoClient.Database("mooshroombase").Collection("users")

	user, err := utils.FindUserFromMongoDBUsingID(userId, coll)

	if err != nil {
		c.WriteJSON(types.ErrorResponse{Error: "User Not Found!"})
		c.Close()
		return
	}

	c.WriteJSON(types.HttpSuccessResponse{Data: map[string]any{"userData": user}})

	cur, err := coll.Watch(context.TODO(), mongo.Pipeline{})
	if err != nil {
		c.WriteJSON(types.ErrorResponse{Error: "Failed to establish change stream: " + err.Error()})
		c.Close()
		return
	}

	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {
		user, err = utils.FindUserFromMongoDBUsingID(user.ID, coll)
		if err != nil {
			c.WriteJSON("Something Went Wrong")
			c.Close()
			return
		}
		c.WriteJSON(types.HttpSuccessResponse{Data: map[string]any{"user": user}})
	}
}
