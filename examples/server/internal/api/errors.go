package api

import (
	"context"
	"net/http"

	"github.com/Kegian/agen/examples/server/internal/generated/oapi"

	"github.com/Kegian/agen"
)

var (
	ErrUnexcpected = agen.AddError(10, http.StatusBadRequest, "unexpected something")
)

func (s *Service) NewError(_ context.Context, err error) *oapi.ErrorStatusCode {
	e := agen.GetErrorInfo(err)
	return &oapi.ErrorStatusCode{
		StatusCode: e.StatusCode,
		Response: oapi.Error{
			Code:    e.Code,
			Message: e.Message,
			Debug:   oapi.OptString{Value: e.Debug, Set: e.Debug != ""},
		},
	}
}
