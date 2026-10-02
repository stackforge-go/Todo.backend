package users

import (
	"time"

	"github.com/google/uuid"
	"github.com/stackforge-go/Todo.backend/internal/infrastructure/errs"
	"github.com/stackforge-go/Todo.backend/internal/transport/http"
)

type httpHandler struct {
	svc Service
}

func NewHTTPHandler(svc Service) *httpHandler {
	return &httpHandler{
		svc: svc,
	}
}

func (h *httpHandler) Routes() []http.Route {
	return []http.Route{
		{
			Method:  http.GET,
			Path:    "/users",
			Handler: h.List,
		},
		{
			Method:  http.GET,
			Path:    "/users/{id}",
			Handler: h.FindByID,
		},
	}
}

// ============================================================
// DTO
// ============================================================

// UserResponse — публичное представление пользователя.
// Используется в users и auth.
type UserResponse struct {
	ID            uuid.UUID `json:"id"`
	EmailVerified bool      `json:"emailVerified"`
	Email         string    `json:"email"`
	FullName      string    `json:"fullName,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// ToUserResponse конвертирует доменное представление в DTO.
func ToUserResponse(u PublicUser) UserResponse {
	return UserResponse{
		ID:            u.ID,
		EmailVerified: u.EmailVerified,
		Email:         u.Email,
		FullName:      u.FullName,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
}

// ============================================================
// FindByID
// ============================================================

type findByIDResponse struct {
	User UserResponse `json:"user"`
}

func (h *httpHandler) FindByID(ctx *http.Context) {
	id, err := uuid.Parse(ctx.Request().Param("id"))
	if err != nil {
		ctx.Error(errs.InvalidArgument.
			WithMessage("invalid id").
			WithOp("http.users.FindByID").
			Wrap(err))
		return
	}

	out, err := h.svc.FindByID(ctx.Context(), FindByIDParams{ID: id})
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.Response().OK(findByIDResponse{
		User: ToUserResponse(out.User),
	})
}

// ============================================================
// List
// ============================================================

type listResponse struct {
	Users  []UserResponse `json:"users"`
	Total  int64          `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

func (h *httpHandler) List(ctx *http.Context) {
	limit, err := ctx.Request().QueryIntDefault("limit", 0)
	if err != nil {
		ctx.Error(errs.InvalidArgument.
			WithMessage("invalid limit").
			WithOp("http.users.List").
			Wrap(err))
		return
	}

	offset, err := ctx.Request().QueryIntDefault("offset", 0)
	if err != nil {
		ctx.Error(errs.InvalidArgument.
			WithMessage("invalid offset").
			WithOp("http.users.List").
			Wrap(err))
		return
	}

	out, err := h.svc.List(ctx.Context(), ListParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		ctx.Error(err)
		return
	}

	users := make([]UserResponse, 0, len(out.Users))
	for _, u := range out.Users {
		users = append(users, ToUserResponse(u))
	}

	ctx.Response().OK(listResponse{
		Users:  users,
		Total:  out.Total,
		Limit:  out.Limit,
		Offset: out.Offset,
	})
}
