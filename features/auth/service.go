package auth

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"time"

	"example.com/inga/internal/database/models"
	platformErrors "example.com/inga/internal/errors"
	"example.com/inga/internal/logger"
	"example.com/inga/internal/notification/emails"
	"example.com/inga/internal/notification/queue"
	"example.com/inga/internal/utils/generator"
	"example.com/inga/internal/utils/password"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Service interface {
	registerUserHandler(ctx *gin.Context, params SignupParams) (SignupResponse, error)
	VerifyToken(ctx context.Context, token string) error
	UpdateUsername(ctx context.Context, params UpdateUsernameParams) (UpdateUsernameResponse, error)
	authenticateUserHandler(ctx *gin.Context, params SignupParams) (LoginResponse, error)
}

type service struct {
	db             models.Querier
	emailQueue     *queue.EmailQueue
	verifyURL      string
	tokenGenerator generator.TokenGenerator
}

func NewService(db models.Querier, emailQueue *queue.EmailQueue, verifyURL string, tokenGenerator generator.TokenGenerator) Service {
	return &service{
		db:             db,
		emailQueue:     emailQueue,
		verifyURL:      verifyURL,
		tokenGenerator: tokenGenerator,
	}
}

func (s service) registerUserHandler(ctx *gin.Context, params SignupParams) (SignupResponse, error) {
	user, err := s.db.GetUserByEmail(ctx, params.Email)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			logger.Error(ctx, "failed to get user by email", zap.Error(err))
			return SignupResponse{}, platformErrors.ErrInternal
		}
	}
	if user.Email != "" {
		return SignupResponse{}, platformErrors.MakeApiError(400, "user with this email already exists")
	}

	newUser, err := s.db.SaveUser(ctx, models.SaveUserParams{
		Email:    params.Email,
		Password: password.DefaultHasher.Hash(params.Password),
	})

	if err != nil {
		logger.Error(ctx, "failed to save user", zap.Error(err))
		return SignupResponse{}, platformErrors.ErrInternal
	}

	if s.emailQueue != nil {
		token, err := generator.GenerateVerificationToken(newUser.ID)
		if err != nil {
			logger.Error(ctx, "failed to generate verification token", zap.Error(err))
		} else {
			expiresAt := time.Now().Add(24 * time.Hour)
			if _, err := s.db.SaveVerification(ctx, models.SaveVerificationParams{
				UserID:    newUser.ID,
				Token:     token,
				ExpiresAt: expiresAt,
			}); err != nil {
				logger.Error(ctx, "failed to save verification token", zap.Error(err))
			} else {
				verifyLink := s.verifyURL + url.QueryEscape(token)
				// Queue email for asynchronous sending
				if err := s.emailQueue.SendTemplate(ctx, emails.Message{
					To:       []string{params.Email},
					Subject:  "Verify your email",
					Template: "verify.html",
					Data: map[string]any{
						"Email":     params.Email,
						"AppName":   "Inga",
						"VerifyURL": verifyLink,
					},
				}); err != nil {
					logger.Error(ctx, "failed to queue verification email", zap.Error(err))
				}
			}
		}
	}

	return SignupResponse{
		UserId:  newUser.ID,
		Message: "Your account has been created. Check your inbox for a verification email.",
	}, nil

}

func (s service) VerifyToken(ctx context.Context, token string) error {
	if token == "" {
		return platformErrors.MakeApiError(400, "token is required")
	}

	v, err := s.db.GetVerificationByToken(ctx, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return platformErrors.MakeApiError(400, "invalid token")
		}
		logger.Error(ctx, "failed to get verification token", zap.Error(err))
		return platformErrors.ErrInternal
	}

	if v.Used.Valid && v.Used.Bool {
		return platformErrors.MakeApiError(400, "token already used")
	}

	if time.Now().After(v.ExpiresAt) {
		return platformErrors.MakeApiError(400, "token expired")
	}

	if err := s.db.MarkVerificationUsed(ctx, token); err != nil {
		logger.Error(ctx, "failed to mark verification token used", zap.Error(err))
		return platformErrors.ErrInternal
	}

	user, err := s.db.GetUserById(ctx, v.UserID)
	if err != nil {
		logger.Error(ctx, "failed to get user by id", zap.Error(err))
		return platformErrors.ErrInternal
	}

	if _, err := s.db.VerifyUser(ctx, user.Email); err != nil {
		logger.Error(ctx, "failed to verify user", zap.Error(err))
		return platformErrors.ErrInternal
	}

	return nil
}

func (s service) UpdateUsername(ctx context.Context, params UpdateUsernameParams) (UpdateUsernameResponse, error) {
	// Check if username already exists
	existingUser, err := s.db.GetUserByUsername(ctx, params.Username)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			logger.Error(ctx, "failed to check username availability", zap.Error(err))
			return UpdateUsernameResponse{}, platformErrors.ErrInternal
		}
	}
	if existingUser.ID != params.UserId && existingUser.Username != "" {
		return UpdateUsernameResponse{}, platformErrors.MakeApiError(400, "username is already taken")
	}

	// Update username
	user, err := s.db.UpdateUsername(ctx, models.UpdateUsernameParams{
		Username: params.Username,
		ID:       params.UserId,
	})
	if err != nil {
		logger.Error(ctx, "failed to update username", zap.Error(err))
		return UpdateUsernameResponse{}, platformErrors.ErrInternal
	}

	return UpdateUsernameResponse{
		Username: user.Username,
		Message:  "username updated successfully",
	}, nil
}

func (s service) authenticateUserHandler(ctx *gin.Context, params SignupParams) (LoginResponse, error) {
	user, err := s.db.GetUserByEmail(ctx, params.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return LoginResponse{}, platformErrors.MakeApiError(401, "invalid email or password")
		}
		logger.Error(ctx, "failed to get user by email", zap.Error(err))
		return LoginResponse{}, platformErrors.ErrInternal
	}

	if !password.DefaultHasher.Validate(params.Password, user.Password) {
		return LoginResponse{}, platformErrors.MakeApiError(401, "invalid email or password")
	}

	if user.ID == uuid.Nil {
		return LoginResponse{}, platformErrors.MakeApiError(401, "invalid email or password")
	}

	tokenData := generator.TokenData{UserID: user.ID}
	token, err := s.tokenGenerator.Generate(tokenData)
	if err != nil {
		logger.Error(ctx, "failed to generate token", zap.Error(err))
		return LoginResponse{}, platformErrors.ErrInternal
	}

	return LoginResponse{
		UserId: user.ID,
		AccessToken: TokenResponse{
			Token:     token.Token,
			ExpiresAt: token.ExpiresAt,
		},
	}, nil

}
