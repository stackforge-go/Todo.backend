package users

import (
	"time"

	"github.com/google/uuid"

	"github.com/stackforge-go/Todo.backend/internal/infrastructure/errs"
	"github.com/stackforge-go/Todo.backend/internal/transport/http"
)

type httpHandler struct {
	uc Usecase
}

func NewHTTPHandler(uc Usecase) *httpHandler {
	return &httpHandler{uc: uc}
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

// userResponse — публичное представление пользователя.
//
// Не содержит PasswordHash и Version. Используется в users и auth,
// чтобы формат user во всех эндпоинтах был одинаков.
type userResponse struct {
	ID            uuid.UUID `json:"id"`
	EmailVerified bool      `json:"emailVerified"`
	Email         string    `json:"email"`
	FullName      string    `json:"fullName,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// ToUserResponse конвертирует доменное представление в DTO.
func ToUserResponse(u PublicUser) userResponse {
	return userResponse{
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

// findByIDResponse — обёртка {user: {...}} для согласованности
// с auth и другими эндпоинтами.
type findByIDResponse struct {
	User userResponse `json:"user"`
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

	out, err := h.uc.GetByID(ctx.Context(), GetByIDParams{ID: id})
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

// listResponse — формат ответа списка.
//
// Users + Total + Limit/Offset — фактически применённые
// (после нормализации в usecase).
type listResponse struct {
	Users  []userResponse `json:"users"`
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

	out, err := h.uc.List(ctx.Context(), ListParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		ctx.Error(err)
		return
	}

	users := make([]userResponse, 0, len(out.Users))
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
