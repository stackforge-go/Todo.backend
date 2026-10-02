package auth

import (
	"errors"

	"github.com/stackforge-go/Todo.backend/internal/infrastructure/errs"
	"github.com/stackforge-go/Todo.backend/internal/infrastructure/token"
)

// ============================================================
// Константы
// ============================================================

// op* — операции для логов. Используются в WithOp,
// чтобы в логах был op="auth.Register", а не голый error.
const (
	opRegister = "auth.Register"
	opLogin    = "auth.Login"
	opRefresh  = "auth.Refresh"
)

// ============================================================
// Реализация
// ============================================================

// usecase — реализация Usecase. Приватная.
//
// Зависимости — только интерфейсы (UserProvider, Hasher, TokenIssuer).
// Конкретные реализации (users.Usecase, BcryptHasher, JWTIssuer)
// подставляются в main.go.
type usecase struct {
	users  UserProvider
	tokens token.Issuer
}

func NewUsecase(
	users UserProvider,
	tokens token.Issuer,
) *usecase {
	return &usecase{
		users:  users,
		tokens: tokens,
	}
}

// Compile-time проверка: *usecase реализует Usecase.
var _ Usecase = (*usecase)(nil)

// ============================================================
// Маппинг доменных ошибок
// ============================================================

// mapDomainError конвертирует доменные ошибки auth в errs.AppError.
//
// Домен не знает про errs — это ответственность usecase.
func mapDomainError(err error) *errs.AppError {
	switch {
	case errors.Is(err, ErrInvalidCredentials):
		return errs.Unauthorized.WithMessage("invalid credentials").Wrap(err)

	case errors.Is(err, ErrInvalidToken):
		return errs.Unauthorized.WithMessage("invalid token").Wrap(err)

	case errors.Is(err, ErrEmailNotVerified):
		return errs.Forbidden.WithMessage("email not verified").Wrap(err)

	default:
		return errs.Internal.WithMessage("internal error").Wrap(err)
	}
}
