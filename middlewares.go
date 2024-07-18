package agen

import (
	"fmt"
	"runtime/debug"

	"github.com/Kegian/agen/errors"
	"github.com/ogen-go/ogen/middleware"
	"go.uber.org/zap"
)

type NextFunc = func(req middleware.Request) (middleware.Response, error)

func RecoverMiddleware(req middleware.Request, next NextFunc) (res middleware.Response, err error) {
	defer func() {
		err2 := recover()
		if err2 != nil {
			SentryRecover(req.Context, err2)
			zap.L().Error("panic acquired", zap.String("stack", string(debug.Stack())))
			err = errors.New("internal server error")
		}
	}()

	res, err = next(req)
	return res, err
}

func LogMiddleware(keys ...any) middleware.Middleware {
	return func(req middleware.Request, next NextFunc) (middleware.Response, error) {
		logger, tags := ctxLoger(req, keys...)
		logger.Info("handling request")

		resp, err := next(req)
		if err != nil {
			logger.Error("failed request", zap.Error(err))
			SentryCaptureException(req.Context, err, tags)
		} else {
			logger.Info("successfull request", zap.Int("status_code", getStatusCode(resp)))
		}

		return resp, err
	}
}

func getStatusCode(res middleware.Response) int {
	if tresp, ok := res.Type.(interface{ GetStatusCode() int }); ok {
		return tresp.GetStatusCode()
	}
	return 0
}

func ctxLoger(req middleware.Request, keys ...any) (*zap.Logger, map[string]string) {
	tags := map[string]string{}
	fields := make([]zap.Field, 0, len(keys)+2)
	fields = append(fields, zap.String("operation", req.OperationName))
	fields = append(fields, zap.String("operation_id", req.OperationID))

	for _, k := range keys {
		key := fmt.Sprintf("%v", k)
		value := fmt.Sprintf("%v", req.Context.Value(k))
		tags[key] = value
		fields = append(fields, zap.String(key, value))
	}

	return zap.L().With(fields...), tags
}
