package utils

import (
	"errors"
	"net/url"
	"strings"
	"sync"
	"unicode"

	"example.com/inga/internal/api"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var registerPasswordValidator sync.Once

type ValidationUtil interface {
	ValidateFields() error
}

func init() {
	registerPasswordValidator.Do(func() {
		if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
			_ = v.RegisterValidation("password", validatePassword)
			_ = v.RegisterValidation("url", validateURL)
		}
	})
}

func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if len(password) < 8 {
		return false
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		case strings.ContainsRune("!@#$%^&*()-_+=[]{}|;:',.<>?/`~\\\"", ch):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecial
}

func validateURL(fl validator.FieldLevel) bool {
	raw := fl.Field().String()
	if raw == "" {
		return false
	}
	u, err := url.ParseRequestURI(raw)
	return err == nil && u.Scheme == "https" && u.Host != ""
}

func ValidateFields(err error) api.Response {
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

		case "URL":
			if fieldErr.Tag() == "required" {
				return api.BadRequest("url is required")
			}
			if fieldErr.Tag() == "url" {
				return api.BadRequest("url is invalid")
			}

		}

		parts := make([]string, 0, len(validationErrs))
		for _, fieldErr := range validationErrs {
			parts = append(parts, fieldErr.Field()+" failed on "+fieldErr.Tag())
		}
		return api.BadRequest(strings.Join(parts, "; "))

	}
	return api.BadRequest("invalid input")
}
