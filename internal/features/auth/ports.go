package auth

import (
	"context"

	"github.com/stackforge-go/Todo.backend/internal/features/users"
	"github.com/stackforge-go/Todo.backend/internal/infrastructure/token"
)

// ============================================================
// Входящий порт
// ============================================================

// Usecase — публичный контракт фичи auth.
//
// Что фича обещает наружу: транспорту, другим фичам.
type Usecase interface {
	Register(ctx context.Context, params RegisterParams) (RegisterResult, error)
	Login(ctx context.Context, params LoginParams) (LoginResult, error)
	Refresh(ctx context.Context, params RefreshParams) (RefreshResult, error)
}

// ============================================================
// Params / Result
// ============================================================

// === Register ===

// RegisterParams — вход Register.
type RegisterParams struct {
	Email    string
	Password string
	FullName string
}

// RegisterResult — выход Register.
//
// User — публичное представление users (без PasswordHash).
// Token — пара токенов (access + refresh).
type RegisterResult struct {
	User    users.PublicUser
	Message string
}

// === Login ===

// LoginParams — вход Login.
type LoginParams struct {
	Email    string
	Password string
}

// LoginResult — выход Login.
type LoginResult struct {
	User   users.PublicUser
	Tokens token.Pair
}

// === Refresh ===

// RefreshParams — вход Refresh.
type RefreshParams struct {
	RefreshToken string
}

// RefreshResult — выход Refresh.
type RefreshResult struct {
	User   users.PublicUser
	Tokens token.Pair
}

// // TokenPair — пара токенов. Возвращается из Issue.
// type TokenPair struct {
// 	AccessToken  string
// 	RefreshToken string
// }

// ============================================================
// Исходящие порты
// ============================================================

// UserProvider — что auth нужно от фичи users.
//
// Узкий интерфейс: только те методы, которые auth реально
// использует. users.Usecase структурно удовлетворяет ему —
// потому что содержит Create с той же сигнатурой.
//
// Использует публичные ТИПЫ users (users.CreateParams,
// users.CreateResult, users.PublicUser). Интерфейс — свой.
type UserProvider interface {
	Create(ctx context.Context, params users.CreateParams) (users.CreateResult, error)
	GetByID(ctx context.Context, params users.GetByIDParams) (users.GetByIDResult, error)
	GetByEmail(ctx context.Context, params users.GetByEmailParams) (users.GetByEmailResult, error)

	// Authenticate — проверка email + пароль.
	// Возвращает PublicUser или ErrInvalidCredentials.
	Authenticate(ctx context.Context, params users.AuthenticateParams) (users.AuthenticateResult, error)
}

// Hasher — проверка пароля.
//
// Отдельный от users.Hasher, потому что auth нужен только Compare.
type Hasher interface {
	Compare(hash, password string) error
}

// // TokenIssuer — выдача и проверка токенов.
// type TokenIssuer interface {
// 	// Issue выдаёт пару токенов для пользователя.
// 	Issue(userID uuid.UUID) (TokenPair, error)

// 	// Validate проверяет refresh-токен и возвращает userID.
// 	Validate(refreshToken string) (uuid.UUID, error)
// }
