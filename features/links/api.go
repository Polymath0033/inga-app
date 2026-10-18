package links

import (
	"example.com/inga/internal/api"
	"example.com/inga/internal/utils"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) SaveLinkHandler(ctx *gin.Context) api.Response {
	var request SaveLinkRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		return utils.ValidateFields(err)
	}

	response, err := h.service.SaveLink(ctx, request)
	if err != nil {
		return api.Error(err)
	}

	return api.Response{
		Code:    200,
		Message: "Link saved successfully",
		Data:    response,
	}
}
