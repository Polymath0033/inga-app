package server

import (
	"example.com/inga/features/auth"
	"example.com/inga/internal/api"
	"example.com/inga/internal/config"
	"example.com/inga/internal/database/models"
	"example.com/inga/internal/storage"
	"github.com/gin-gonic/gin"
)

type Server struct {
	authHandler *auth.Handler
	storage     storage.Storage
	cfg         *config.Config
	db          models.Querier
}

func New(cfg config.Config, db models.Querier, authHandler *auth.Handler, storageSvc storage.Storage) *Server {
	return &Server{
		authHandler: authHandler,
		storage:     storageSvc,
		cfg:         &cfg,
		db:          db,
	}
}

func (s *Server) BuildRoute() (*gin.Engine, error) {
	router := gin.Default()

	// Apply global middleware
	router.Use(CORSMiddleware())
	router.Use(RecoveryMiddleware())
	router.Use(LoggingMiddleware(s.cfg))

	apiGroup := router.Group("/api")

	// Public auth endpoints (no JWT required)
	authGroup := apiGroup.Group("/auth")
	authGroup.POST("/signup", api.Wrap(s.authHandler.SignupHandler))
	authGroup.GET("/verify", api.Wrap(s.authHandler.VerifyHandler))
	authGroup.POST("/login", api.Wrap(s.authHandler.LoginHandler))
	authGroup.POST("/forgot-password", api.Wrap(s.authHandler.ForgotPasswordHandler))
	authGroup.POST("/reset-password", api.Wrap(s.authHandler.ResetPasswordHandler))

	// Protected endpoints (JWT required)
	protectedAuthGroup := authGroup.Group("")
	protectedAuthGroup.Use(NewAuth0Middleware(s.cfg))
	protectedAuthGroup.PATCH("/username", api.Wrap(s.authHandler.UpdateUsernameHandler))

	return router, nil
}
