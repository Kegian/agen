package agen

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	ht "github.com/ogen-go/ogen/http"
	"github.com/ogen-go/ogen/ogenerrors"
	"go.uber.org/zap"
)

type ServerError struct {
	Code       int64
	Message    string
	StatusCode int
}

func (s *ServerError) Error() string {
	return s.Message
}

var AllErrors = map[int64]*ServerError{}

var (
	ErrUnauthorized = AddError(10, 403, "unautorized")
)

func AddError(code int64, statusCode int, msg string) error {
	err := &ServerError{
		Code:       code,
		Message:    msg,
		StatusCode: statusCode,
	}

	if _, ok := AllErrors[code]; ok {
		zap.L().Warn(
			"error code is already present",
			zap.Int64("code", code),
			zap.String("name", msg),
		)
	}

	AllErrors[err.Code] = err

	return err
}

type ErrorInfo struct {
	Code       int64
	Message    string
	Debug      string
	StatusCode int
}

func GetErrorInfo(err error) ErrorInfo {
	var serverErr *ServerError
	if errors.As(err, &serverErr) {
		return ErrorInfo{
			Code:       serverErr.Code,
			Message:    serverErr.Message,
			StatusCode: serverErr.StatusCode,
		}
	}
	return ErrorInfo{
		Code:       0,
		Message:    "internal server error",
		Debug:      err.Error(),
		StatusCode: 500,
	}
}

func ErrorHandler(_ context.Context, w http.ResponseWriter, _ *http.Request, err error) {
	var (
		code    = http.StatusInternalServerError
		ogenErr ogenerrors.Error
	)
	switch {
	case errors.Is(err, ht.ErrNotImplemented):
		code = http.StatusNotImplemented
	case errors.As(err, &ogenErr):
		code = ogenErr.Code()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	body := fmt.Sprintf(`{"code":0,"message":%q}`, err.Error())

	_, _ = io.WriteString(w, body)
}
