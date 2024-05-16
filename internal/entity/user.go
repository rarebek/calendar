// Package entity defines main entities for business logic (services), database mapping and
// HTTP response objects if suitable. Each logic group entities in own file.
package entity

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"regexp"
	"time"
)

type UserRequest struct {
	Email        string `json:"email" example:"nodirbekgolang@gmail.com"`
	Username     string `json:"username"  example:"nodirbek"`
	Password     string `json:"password"  example:"Nodirbek1"`
	RefreshToken string `json:"refresh_token"  example:"nodirbekgolang@gmail.com"`
}

type VerifyResponse struct {
	Id          string    `json:"id"`
	Email       string    `json:"email"`
	Username    string    `json:"username"`
	Password    string    `json:"password"`
	AccessToken string    `json:"access_token"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UserResponse struct {
	Id           string    `json:"id"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	Password     string    `json:"password"`
	RefreshToken string    `json:"refresh_token"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Users struct {
	Users []*UserResponse `json:"users"`
	Count int             `json:"count"`
}

type GetRequest struct {
	Field string `json:"field"`
	Value string `json:"value"`
}

type GetListRequest struct {
	Page    int64  `json:"page"`
	Limit   int64  `json:"limit"`
	OrderBy string `json:"order_by"`
	Field   string `json:"field"`
	Value   string `json:"value"`
}

type UpdateUserRequest struct {
	Id           string `json:"id"`
	Email        string `json:"email"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	RefreshToken string `json:"refresh_token"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

func (ur *UserRequest) IsEmail() error {
	return validation.ValidateStruct(
		ur,
		validation.Field(&ur.Email, validation.Required, is.Email),
	)
}

func (ur *UserRequest) IsComplexPassword() error {
	return validation.Validate(
		&ur.Password,
		validation.Required,
		validation.Length(8, 30),
		validation.Match(regexp.MustCompile("[a-z]|[A-Z][0-9]")),
	)
}
