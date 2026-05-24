package generator

import (
	"context"
	"errors"
	"time"

	"example.com/inga/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenGenerator interface {
	Generate(data TokenData) (TokenResponse, error)
	Validate(token string) (TokenData, error)
}

type TokenData struct {
	UserID    uuid.UUID
	ExpiresAt time.Time
}

type Claim struct {
	TokenData
	jwt.RegisteredClaims
}

func (claim Claim) Validate(ctx context.Context) error {
	return nil
}

type tokenGenerator struct {
	cfg config.JWTConfig
}

func NewTokenGenerator(cfg config.JWTConfig) TokenGenerator {
	return &tokenGenerator{
		cfg: cfg,
	}
}

func (t *tokenGenerator) Generate(data TokenData) (TokenResponse, error) {
	claims := Claim{
		TokenData: data,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   data.UserID.String(),
			Issuer:    t.cfg.Issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(t.cfg.Expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Audience:  []string{t.cfg.Audience},
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(t.cfg.Secret))
	if err != nil {
		return TokenResponse{}, err
	}
	return TokenResponse{
		Token:     tokenString,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(t.cfg.Expiry)).Time,
	}, nil
}

func (t *tokenGenerator) Validate(token string) (TokenData, error) {
	parsedToken, err := jwt.ParseWithClaims(token, &Claim{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(t.cfg.Secret), nil
	})

	if err != nil {
		return TokenData{}, err
	}
	claims, ok := parsedToken.Claims.(*Claim)
	if !ok {
		return TokenData{}, errors.New("invalid token")
	}

	data := claims.TokenData

	return TokenData{
		UserID:    data.UserID,
		ExpiresAt: data.ExpiresAt,
	}, nil

}

type TokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}
