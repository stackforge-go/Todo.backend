package http

import (
	"net/http"

	"github.com/stackforge-go/Todo.backend/internal/infrastructure/errs"
)

// statusFromCode — единственное место в HTTP-транспорте,
// где доменный код маппится в HTTP-статус.
func statusFromCode(code errs.Code) int {
	switch code {
	case errs.CodeInvalidArgument:
		return http.StatusBadRequest
	case errs.CodeUnauthorized:
		return http.StatusUnauthorized
	case errs.CodeForbidden:
		return http.StatusForbidden
	case errs.CodeNotFound:
		return http.StatusNotFound
	case errs.CodeAlreadyExists, errs.CodeConflict:
		return http.StatusConflict
	case errs.CodeRateLimited:
		return http.StatusTooManyRequests
	case errs.CodeTimeout:
		return http.StatusGatewayTimeout
	case errs.CodeUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// ErrorResponse — публичный формат ошибки для клиента.
//
//	code    — машиночитаемый идентификатор (для switch на клиенте)
//	message — человекочитаемое описание (для показа пользователю)
type ErrorResponse struct {
	Code    string `json:"code"    example:"NOT_FOUND"`
	Message string `json:"message" example:"user not found"`
}

// writeError формирует ответ на ошибку. Не логирует —
// логирование живёт в LoggingMiddleware.
func writeError(ctx *Context, err error) {
	appErr, ok := errs.AsAppError(err)
	if !ok {
		appErr = errs.Internal.
			WithMessage("internal error").
			Wrap(err)
	}

	ctx.setAppError(appErr)

	status := statusFromCode(appErr.Code)

	msg := appErr.Message
	if status >= 500 {
		msg = "internal error"
	}

	ctx.Response().JSON(status, ErrorResponse{
		Code:    string(appErr.Code),
		Message: msg,
	})
}
