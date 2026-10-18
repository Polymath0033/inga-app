package auth

import (
	"errors"
	"net/http"
	"strings"

	"example.com/inga/internal/api"
	"example.com/inga/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) SignupHandler(ctx *gin.Context) api.Response {
	var params SignupParams
	if err := ctx.ShouldBindJSON(&params); err != nil {
		return utils.ValidateFields(err)
	}

	response, err := h.service.registerUserHandler(ctx, params)
	if err != nil {
		return api.Error(err)
	}

	return api.Response{
		Code:    http.StatusCreated,
		Message: "Your account has been created. Check your inbox for a verification email.",
		Data:    response,
	}
}

func (h *Handler) LoginHandler(ctx *gin.Context) api.Response {
	var params SignupParams

	if err := ctx.ShouldBindJSON(&params); err != nil {
		response := utils.ValidateFields(err)
		//
		return api.Error(errors.New(response.Message))
	}

	response, err := h.service.authenticateUserHandler(ctx, params)
	if err != nil {
		return api.Error(err)
	}

	return api.Response{
		Code:    http.StatusOK,
		Message: "Login successful",
		Data:    response,
	}

}

func (h *Handler) ForgotPasswordHandler(ctx *gin.Context) api.Response {
	var params EmailRequest
	if err := ctx.ShouldBindJSON(&params); err != nil {
		return utils.ValidateFields(err)
	}

	if err := h.service.forgetPassword(ctx.Request.Context(), params.Email); err != nil {
		return api.Error(err)
	}

	return api.OK("password reset link sent to your email", nil)
}

func (h *Handler) ResetPasswordHandler(ctx *gin.Context) api.Response {
	var params ResetPasswordRequest
	if err := ctx.ShouldBindJSON(&params); err != nil {
		return utils.ValidateFields(err)
	}

	response, err := h.service.resetPassword(ctx.Request.Context(), params)
	if err != nil {
		return api.Error(err)
	}
	return api.Response{
		Code:    http.StatusOK,
		Message: "password reset successfully",
		Data:    response,
	}
}

func validationResponse(err error) api.Response {
	var validationErrs validator.ValidationErrors
	if !errors.As(err, &validationErrs) {
		return api.BadRequest(err.Error())
	}

	for _, fieldErr := range validationErrs {
		switch fieldErr.Field() {
		case "Password":
			if fieldErr.Tag() == "required" {
				return api.BadRequest("password is required")
			}
			if fieldErr.Tag() == "password" {
				return api.BadRequest("password must be at least 8 characters and include uppercase, lowercase, number, and special character")
			}
		case "Email":
			if fieldErr.Tag() == "required" {
				return api.BadRequest("email is required")
			}
			if fieldErr.Tag() == "email" {
				return api.BadRequest("email must be a valid email address")
			}
		case "Username":
			if fieldErr.Tag() == "required" {
				return api.BadRequest("username is required")
			}
			if fieldErr.Tag() == "min" {
				return api.BadRequest("username must be at least 3 characters long")
			}
			if fieldErr.Tag() == "max" {
				return api.BadRequest("username must not exceed 255 characters")
			}
		}
	}

	parts := make([]string, 0, len(validationErrs))
	for _, fieldErr := range validationErrs {
		parts = append(parts, fieldErr.Field()+" failed on "+fieldErr.Tag())
	}
	return api.BadRequest(strings.Join(parts, "; "))
}

func (h *Handler) VerifyHandler(ctx *gin.Context) api.Response {
	token := ctx.Query("token")
	if token == "" {
		return api.BadRequest("token is required")
	}

	if err := h.service.VerifyToken(ctx.Request.Context(), token); err != nil {
		return api.Error(err)
	}

	return api.OK("email verified successfully", nil)
}

func (h *Handler) UpdateUsernameHandler(ctx *gin.Context) api.Response {
	var params UpdateUsernameParams
	if err := ctx.ShouldBindJSON(&params); err != nil {
		return utils.ValidateFields(err)
	}

	response, err := h.service.UpdateUsername(ctx.Request.Context(), params)
	if err != nil {
		return api.Error(err)
	}

	return api.Response{
		Code:    http.StatusOK,
		Message: "username updated successfully",
		Data:    response,
	}
}
