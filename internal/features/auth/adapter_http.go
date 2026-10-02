package auth

import (
	"time"

	"github.com/google/uuid"

	"github.com/stackforge-go/Todo.backend/internal/features/users"
	"github.com/stackforge-go/Todo.backend/internal/infrastructure/errs"
	"github.com/stackforge-go/Todo.backend/internal/transport/http"
)

// ============================================================
// Cookie
// ============================================================

const (
	// refreshCookieName — имя cookie для refresh-токена.
	refreshCookieName = "refresh_token"

	// refreshCookiePath — cookie отправляется только на /api/v1/auth/*.
	// Не на весь API — меньше шума в запросах.
	refreshCookiePath = "/api/v1/auth"

	// refreshCookieTTL — время жизни cookie. Должно совпадать
	// с TTL refresh-токена на стороне Issuer.
	refreshCookieTTL = 7 * 24 * time.Hour
)

// ============================================================
// Handler
// ============================================================

type httpHandler struct {
	uc           Usecase
	cookieSecure bool // true в проде (HTTPS)
	cookieDomain string
}

type HandlerOption func(*httpHandler)

// WithCookieSecure — Secure-флаг для cookie. true в проде.
func WithCookieSecure(secure bool) HandlerOption {
	return func(h *httpHandler) {
		h.cookieSecure = secure
	}
}

// WithCookieDomain — Domain для cookie.
func WithCookieDomain(domain string) HandlerOption {
	return func(h *httpHandler) {
		h.cookieDomain = domain
	}
}

func NewHTTPHandler(uc Usecase, opts ...HandlerOption) *httpHandler {
	h := &httpHandler{
		uc:           uc,
		cookieSecure: true,
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

func (h *httpHandler) Routes() []http.Route {
	return []http.Route{
		{Method: http.POST, Path: "/auth/register", Handler: h.Register},
		{Method: http.POST, Path: "/auth/login", Handler: h.Login},
		{Method: http.POST, Path: "/auth/refresh", Handler: h.Refresh},
		{Method: http.POST, Path: "/auth/logout", Handler: h.Logout},
	}
}

// ============================================================
// DTO
// ============================================================

type userResponse struct {
	ID            uuid.UUID `json:"id"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"emailVerified"`
	FullName      string    `json:"fullName,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func toUserResponse(u users.PublicUser) userResponse {
	return userResponse{
		ID:            u.ID,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
		FullName:      u.FullName,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
}

// registerResponse — user + message. Токенов нет.
type registerResponse struct {
	User    userResponse `json:"user"`
	Message string       `json:"message"`
}

// authResponse — user + accessToken. RefreshToken НЕ в JSON —
// он уходит в HttpOnly cookie.
type authResponse struct {
	User        userResponse `json:"user"`
	AccessToken string       `json:"accessToken"`
}

// ============================================================
// Register
// ============================================================

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"fullName"`
}

func (h *httpHandler) Register(ctx *http.Context) {
	var req registerRequest
	if err := ctx.Request().BindJSON(&req); err != nil {
		ctx.Error(errs.InvalidArgument.
			WithMessage("invalid JSON body").
			WithOp("http.auth.Register").
			Wrap(err))
		return
	}

	out, err := h.uc.Register(ctx.Context(), RegisterParams{
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
	})
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.Response().Header().Set("Location", "/api/v1/users/"+out.User.ID.String())
	ctx.Response().Created(registerResponse{
		User:    toUserResponse(out.User),
		Message: out.Message,
	})
}

// ============================================================
// Login
// ============================================================

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *httpHandler) Login(ctx *http.Context) {
	var req loginRequest
	if err := ctx.Request().BindJSON(&req); err != nil {
		ctx.Error(errs.InvalidArgument.
			WithMessage("invalid JSON body").
			WithOp("http.auth.Login").
			Wrap(err))
		return
	}

	out, err := h.uc.Login(ctx.Context(), LoginParams{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		ctx.Error(err)
		return
	}

	// Refresh — в HttpOnly cookie. Access — в JSON.
	h.setRefreshCookie(ctx, out.Tokens.RefreshToken)

	ctx.Response().OK(authResponse{
		User:        toUserResponse(out.User),
		AccessToken: out.Tokens.AccessToken,
	})
}

// ============================================================
// Refresh
// ============================================================

func (h *httpHandler) Refresh(ctx *http.Context) {
	// Refresh-токен — только из cookie. Из body не принимаем.
	refreshToken := ctx.Request().Cookie(refreshCookieName)
	if refreshToken == "" {
		ctx.Error(errs.Unauthorized.
			WithMessage("missing refresh token").
			WithOp("http.auth.Refresh"))
		return
	}

	out, err := h.uc.Refresh(ctx.Context(), RefreshParams{
		RefreshToken: refreshToken,
	})
	if err != nil {
		// Инвалидный refresh — удаляем cookie.
		h.clearRefreshCookie(ctx)
		ctx.Error(err)
		return
	}

	// Ротация refresh-токена: новый refresh — в cookie.
	h.setRefreshCookie(ctx, out.Tokens.RefreshToken)

	ctx.Response().OK(authResponse{
		User:        toUserResponse(out.User),
		AccessToken: out.Tokens.AccessToken,
	})
}

// ============================================================
// Logout
// ============================================================

func (h *httpHandler) Logout(ctx *http.Context) {
	h.clearRefreshCookie(ctx)
	ctx.Response().NoContent()
}

// ============================================================
// Cookie helpers
// ============================================================

// setRefreshCookie ставит refresh-токен в HttpOnly cookie.
//
// Флаги:
//   - HttpOnly — JS не может прочитать (защита от XSS).
//   - Secure   — только по HTTPS (в проде true).
//   - SameSite=Strict — CSRF-защита.
//   - Path     — только на /api/v1/auth/*.
//   - MaxAge   — время жизни refresh-токена.
func (h *httpHandler) setRefreshCookie(ctx *http.Context, token string) {
	cookie := &http.Cookie{
		Name:     refreshCookieName,
		Value:    token,
		Path:     refreshCookiePath,
		Domain:   h.cookieDomain,
		MaxAge:   int(refreshCookieTTL.Seconds()),
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteStrictMode,
	}

	ctx.Response().SetCookie(cookie)
}

// clearRefreshCookie удаляет refresh-cookie (logout, invalid token).
//
// MaxAge: -1 — браузер удалит cookie немедленно.
func (h *httpHandler) clearRefreshCookie(ctx *http.Context) {
	cookie := &http.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     refreshCookiePath,
		Domain:   h.cookieDomain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteStrictMode,
	}

	ctx.Response().SetCookie(cookie)
}
