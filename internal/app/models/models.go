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
	OrderNumber string `json:"order"`
	Sum         int    `json:"sum"`
}

type Order struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    int       `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
	UserID     int       `json:"-"`
}

type Balance struct {
	Current   int `json:"current"`
	Withdrawn int `json:"withdrawn"`
}

type Withdraw struct {
	OrderNumber string    `json:"order"`
	Sum         int       `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

type User struct {
	ID       int
	Login    string
	Password string
}

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}
