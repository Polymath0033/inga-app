package auth

import (
	"time"

	"github.com/google/uuid"
)

type SignupParams struct {
	Password string `json:"password" binding:"required,password"`
	Email    string `json:"email" binding:"required,email"`
}

type SignupResponse struct {
	UserId uuid.UUID `json:"userId"`
}

type UpdateUsernameParams struct {
	UserId   uuid.UUID `json:"userId" binding:"required"`
	Username string    `json:"username" binding:"required,min=3,max=255"`
}

type UpdateUsernameResponse struct {
	Username string `json:"username"`
}

type LoginResponse struct {
	UserId      uuid.UUID     `json:"userId"`
	AccessToken TokenResponse `json:"accessToken"`
}

type TokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type ResetPasswordRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,password"`
}

type ResetPasswordResponse struct {
	Message string `json:"message"`
}

type EmailRequest struct {
	Email string `json:"email" binding:"required,email"`
}
