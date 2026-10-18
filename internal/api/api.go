package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int
	Message string
	Data    interface{}
	Error   *ApiError
}

func (r *Response) Marshal() ([]byte, error) {
	errorMessage := ""
	if r.Error != nil {
		errorMessage = r.Error.Message
	} else if r.Code >= http.StatusBadRequest {
		errorMessage = r.Message
	}

	data, err := json.Marshal(ResponseEnvelope{
		Code:    r.Code,
		Message: r.Message,
		Data:    r.Data,
		Error:   errorMessage,
	})
	if err != nil {
		return nil, err
	}
	return data, nil
}

type HttHandler func(ctx *gin.Context) Response

func Wrap(handler HttHandler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		response := handler(ctx)
		data, err := response.Marshal()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, ResponseEnvelope{
				Code:    http.StatusInternalServerError,
				Message: "Internal Server Error",
				Data:    nil,
				Error:   "Internal Server Error",
			})
			return
		}
		ctx.Data(response.Code, "application/json", data)
	}
}

type ResponseEnvelope struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Error   string      `json:"error,omitempty"`
}

type ApiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *ApiError) Error() string {
	return e.Message
}

func NewErrorWithCode(code int, message string) *ApiError {
	return &ApiError{
		Code:    code,
		Message: message,
	}
}

func BadRequest(message string) Response {
	return Response{
		Code:    http.StatusBadRequest,
		Message: message,
		Error:   NewErrorWithCode(http.StatusBadRequest, message),
	}
}

func Error(err error) Response {
	var e *ApiError
	if errors.As(err, &e) {
		code := e.Code
		if code == 0 {
			code = http.StatusInternalServerError
		}
		apiErr := *e
		apiErr.Code = code
		return Response{
			Code:    code,
			Message: apiErr.Message,
			Error:   &apiErr,
		}
	}
	return Response{
		Code:    http.StatusInternalServerError,
		Message: err.Error(),
		Error: &ApiError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		},
	}
}

func Unauthorized(message string) Response {
	return Response{
		Code:    http.StatusUnauthorized,
		Message: message,
		Error:   NewErrorWithCode(http.StatusUnauthorized, message),
	}
}

func Forbidden(message string) Response {
	return Response{
		Code:    http.StatusForbidden,
		Message: message,
		Error:   NewErrorWithCode(http.StatusForbidden, message),
	}
}

func NotFound(message string) Response {
	return Response{
		Code:    http.StatusNotFound,
		Message: message,
		Error:   NewErrorWithCode(http.StatusNotFound, message),
	}
}

func ServerError(message string) Response {
	return Response{
		Code:    http.StatusInternalServerError,
		Message: message,
		Error:   NewErrorWithCode(http.StatusInternalServerError, message),
	}
}

func OK(message string, data interface{}) Response {
	return Response{
		Code:    http.StatusOK,
		Message: message,
		Data:    data,
	}
}
