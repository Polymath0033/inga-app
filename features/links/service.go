package links

import (
	"example.com/inga/internal/database/models"
	"example.com/inga/internal/storage"
	"github.com/gin-gonic/gin"
)

type Service interface {
	SaveLink(ctx *gin.Context, params SaveLinkRequest) (SaveLinkResponse, error)
}

type service struct {
	db         models.Querier
	cloudinary storage.Storage
}

func NewService(db models.Querier, cld storage.Storage) Service {
	return &service{
		db:         db,
		cloudinary: cld,
	}
}

func (s service) SaveLink(ctx *gin.Context, params SaveLinkRequest) (SaveLinkResponse, error) {

	return SaveLinkResponse{}, nil

}
