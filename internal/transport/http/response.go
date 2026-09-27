package http

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	w          http.ResponseWriter
	statusCode int
	written    bool
	bytes      int
}

func NewResponse(w http.ResponseWriter) *Response {
	return &Response{
		w:          w,
		statusCode: http.StatusOK,
	}
}

func (r *Response) Raw() http.ResponseWriter {
	return r.w
}

func (r *Response) Header() http.Header {
	return r.w.Header()
}

func (r *Response) Status() int {
	return r.statusCode
}

func (r *Response) Written() bool {
	return r.written
}

func (r *Response) BytesWritten() int {
	return r.bytes
}

func (r *Response) WriteHeader(statusCode int) {
	if r.written {
		return
	}

	r.statusCode = statusCode
	r.written = true

	r.w.WriteHeader(statusCode)
}

func (r *Response) Write(body []byte) (int, error) {
	if !r.written {
		r.WriteHeader(http.StatusOK)
	}

	n, err := r.w.Write(body)
	r.bytes += n

	return n, err
}

func (r *Response) JSON(statusCode int, body any) {
	data, err := json.Marshal(body)
	if err != nil {
		// Сюда мы попадаем только если сериализация body упала.
		// Отдаём минимально валидный ответ, чтобы клиент не завис.
		http.Error(
			r.w,
			`{"error":"internal error","code":"INTERNAL"}`,
			http.StatusInternalServerError,
		)
		return
	}

	r.Header().Set("Content-Type", "application/json")
	r.WriteHeader(statusCode)
	_, _ = r.Write(data)
}

func (r *Response) OK(body any) {
	r.JSON(http.StatusOK, body)
}

func (r *Response) Created(body any) {
	r.JSON(http.StatusCreated, body)
}

func (r *Response) NoContent() {
	r.WriteHeader(http.StatusNoContent)
}
