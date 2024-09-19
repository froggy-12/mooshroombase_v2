package types

import (
	"database/sql"
	"time"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type HttpSuccessResponse struct {
	Message string         `json:"message"`
	Data    map[string]any `json:"data"`
}

type SingleFileUploadedSuccessResponse struct {
	FileName string `json:"fileName"`
	Message  string `json:"message"`
}

type MultipleFileUploadedSuccessResponse struct {
	FileNames []string `json:"fileNames"`
	Message   string   `json:"message"`
}

type DeleteSuccessResponse struct {
	FileName string `json:"fileName"`
	Message  string `json:"message"`
}

type User struct {
	ID             string `json:"id"`
	UserName       string `json:"username" validate:"required"`
	FirstName      string `json:"firstName" validate:"required"`
	LastName       string `json:"lastName" validate:"required"`
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required,min=8"`
	ProfilePicture string `json:"profilePicture"`
}

type UpdateMongoUser struct {
	FirstName      string        `json:"firstName"`
	LastName       string        `json:"lastName"`
	ProfilePicture string        `json:"profilePicture"`
	RawData        []RawUserData `json:"rawData"`
}

type UpdateMariaUser struct {
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	ProfilePicture string `json:"profilePicture"`
}

type UpdateMongoUserRawData struct {
	RawData RawUserData `json:"rawData"`
}

type User_Mongo struct {
	ID                string           `bson:"id"`
	UserName          string           `bson:"username, unique"`
	FirstName         string           `bson:"firstName"`
	LastName          string           `bson:"lastName"`
	Email             string           `bson:"email, unique"`
	Password          string           `bson:"password"`
	ProfilePicture    string           `bson:"profilePicture"`
	CreatedAt         time.Time        `bson:"createdAt"`
	UpdatedAt         time.Time        `bson:"updatedAt"`
	Verified          bool             `bson:"verified"`
	VerificationToken string           `bson:"verificationToken"`
	LastLoggedIn      LastTimeLoggedIn `bson:"lastLoggedIn"`
	RawData           []RawUserData    `bson:"rawData"`
}

type LastTimeLoggedIn struct {
	When time.Time `bson:"when"`
}

type RawUserData struct {
	Data map[string]any `bson:"data"`
}

type User_Maria struct {
	ID                string
	UserName          string
	FirstName         string
	LastName          string
	Email             string
	Password          string
	ProfilePicture    string
	CreatedAt         sql.NullTime
	UpdatedAt         sql.NullTime
	Verified          bool
	VerificationToken string
	LastLoggedIn      sql.NullTime
}

type User_Maria_Oauth struct {
	ID                string
	UserName          string
	FirstName         string
	LastName          string
	Email             string
	ProfilePicture    string
	CreatedAt         sql.NullTime
	UpdatedAt         sql.NullTime
	Verified          bool
	OAuthProvider     string
	VerificationToken string
}

type User_Maria_Oauth_Payload struct {
	ID             string `json:"id"`
	UserName       string `json:"username"`
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	Email          string `json:"email"`
	ProfilePicture string `json:"profilePicture"`
	Verified       bool   `json:"verified"`
	OAuthProvider  string `json:"oauthProvider"`
}

type LogInDetails struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}
