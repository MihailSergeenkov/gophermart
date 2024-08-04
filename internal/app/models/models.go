package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type RegisterUserRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type RegisterUserResponse struct {
	AuthToken string
}

type LoginUserRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginUserResponse struct {
	AuthToken string
}

type AddWithdrawRequest struct {
	OrderNumber string  `json:"order"`
	Sum         float32 `json:"sum"`
}

type Order struct {
	UploadedAt time.Time `json:"uploaded_at"`
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float32   `json:"accrual,omitempty"`
	UserID     int       `json:"-"`
}

type Balance struct {
	Current   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

type Withdraw struct {
	ProcessedAt time.Time `json:"processed_at"`
	OrderNumber string    `json:"order"`
	Sum         float32   `json:"sum"`
}

type User struct {
	Login    string
	Password []byte
	ID       int
}

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}
