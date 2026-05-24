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
	UserId  uuid.UUID `json:"userId"`
	Message string    `json:"message"`
}

type UpdateUsernameParams struct {
	UserId   uuid.UUID `json:"userId" binding:"required"`
	Username string    `json:"username" binding:"required,min=3,max=255"`
}

type UpdateUsernameResponse struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

type LoginResponse struct {
	UserId      uuid.UUID     `json:"userId"`
	AccessToken TokenResponse `json:"accessToken"`
}

type TokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}
