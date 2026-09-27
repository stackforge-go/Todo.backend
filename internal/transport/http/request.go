package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var validate = validator.New()

type Request struct {
	r *http.Request
}

func NewRequest(r *http.Request) *Request {
	return &Request{r: r}
}

func (r *Request) Context() context.Context {
	return r.r.Context()
}

func (r *Request) SetContext(ctx context.Context) {
	*r.r = *r.r.WithContext(ctx)
}

func (r *Request) Raw() *http.Request {
	return r.r
}

func (r *Request) Method() Method {
	return Method(r.r.Method)
}

func (r *Request) Path() string {
	return r.r.URL.Path
}

func (r *Request) RemoteAddr() string {
	return r.r.RemoteAddr
}

func (r *Request) Param(name string) string {
	return r.r.PathValue(name)
}

func (r *Request) ParamUUID(name string) (uuid.UUID, error) {
	return uuid.Parse(r.r.PathValue(name))
}

func (r *Request) Query(name string) string {
	return r.r.URL.Query().Get(name)
}

func (r *Request) QueryIntDefault(name string, def int) (int, error) {
	raw := r.Query(name)
	if raw == "" {
		return def, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("query parameter %q must be int: %w", name, err)
	}
	return v, nil
}

// func (r *Request) QueryInt(name string) (*int, error) {
// 	param := r.r.URL.Query().Get(name)
// 	if param == "" {
// 		return nil, nil
// 	}

// 	val, err := strconv.Atoi(param)
// 	if err != nil {
// 		return nil, fmt.Errorf(
// 			"%w", err,
// 		)
// 	}
// 	return strconv.Atoi(r.r.URL.Query().Get(name))
// }

func (r *Request) Header(name string) string {
	return r.r.Header.Get(name)
}

func (r *Request) BindJSON(dest any) error {
	defer r.r.Body.Close()

	if err := json.NewDecoder(r.r.Body).Decode(dest); err != nil {
		return fmt.Errorf("decode JSON: %w", err)
	}

	if err := validate.Struct(dest); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	return nil
}
