package http

import (
	"context"
	"net/http"

	"github.com/stackforge-go/Todo.backend/internal/infrastructure/errs"
	"github.com/stackforge-go/Todo.backend/internal/infrastructure/logger"
)

type Context struct {
	request  *Request
	response *Response
	logger   logger.Logger
	appErr   *errs.AppError
}

func newContext(
	w http.ResponseWriter,
	r *http.Request,
	log logger.Logger,
) *Context {
	return &Context{
		request:  NewRequest(r),
		response: NewResponse(w),
		logger:   log,
	}
}

func (c *Context) Request() *Request {
	return c.request
}

func (c *Context) Response() *Response {
	return c.response
}

func (c *Context) Context() context.Context {
	return c.request.Context()
}

func (c *Context) SetContext(ctx context.Context) {
	c.request.SetContext(ctx)
}

func (c *Context) Logger() logger.Logger {
	return c.logger
}

// Error — единая точка отдачи ошибки из хендлера.
// Внутри формирует ответ и запоминает AppError для логирования.
func (c *Context) Error(err error) {
	writeError(c, err)
}

// setAppError / AppError — внутренний обмен между writeError
// и LoggingMiddleware. Наружу не торчат.
func (c *Context) setAppError(err *errs.AppError) {
	c.appErr = err
}

func (c *Context) AppError() *errs.AppError {
	return c.appErr
}
