package server

import (
	"example.com/inga/features/auth"
	"example.com/inga/internal/api"
	"example.com/inga/internal/config"
	"example.com/inga/internal/database/models"
	"github.com/gin-gonic/gin"
)

type Server struct {
	authHandler *auth.Handler
	cfg         *config.Config
	db          models.Querier
}

func New(cfg config.Config, db models.Querier, authHandler *auth.Handler) *Server {
	return &Server{
		authHandler: authHandler,
		cfg:         &cfg,
		db:          db,
	}
}

func (s *Server) BuildRoute() (*gin.Engine, error) {
	router := gin.Default()

	apiGroup := router.Group("/api")
	authGroup := apiGroup.Group("/auth")
	authGroup.POST("/signup", api.Wrap(s.authHandler.SignupHandler))
	authGroup.GET("/verify", api.Wrap(s.authHandler.VerifyHandler))
	authGroup.PATCH("/username", api.Wrap(s.authHandler.UpdateUsernameHandler))

	return router, nil
}
