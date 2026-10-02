package users

import (
	"context"

	"github.com/google/uuid"
)

// ============================================================
// Входящий порт
// ============================================================

// Usecase — публичный контракт фичи users.
//
// Что фича обещает наружу: другим фичам, транспорту.
// Реализуется *usecase (см. service.go).
type Usecase interface {
	Create(ctx context.Context, params CreateParams) (CreateResult, error)
	GetByID(ctx context.Context, params GetByIDParams) (GetByIDResult, error)
	GetByEmail(ctx context.Context, params GetByEmailParams) (GetByEmailResult, error)
	List(ctx context.Context, params ListParams) (ListResult, error)

	// Authenticate — проверка email + пароль.
	// Возвращает PublicUser (без PasswordHash) при успехе,
	// ErrInvalidCredentials при любой неудаче.
	Authenticate(ctx context.Context, params AuthenticateParams) (AuthenticateResult, error)
}

// ============================================================
// Params / Result
// ============================================================

// === Create ===

// CreateParams — вход Create.
//
// Password — СЫРОЙ (до хеширования). Хеш делает usecase.
// EmailVerified — вызывающий решает: auth → false, admin → true.
type CreateParams struct {
	Email         string
	Password      string
	FullName      string
	EmailVerified bool
}

// CreateResult — выход Create. PublicUser — без PasswordHash.
type CreateResult struct {
	User PublicUser
}

// === GetByID ===

type GetByIDParams struct {
	ID uuid.UUID
}
type GetByIDResult struct {
	User PublicUser
}

// === GetByEmail ===

type GetByEmailParams struct {
	Email string
}
type GetByEmailResult struct {
	User PublicUser
}

// === Authenticate ===

type AuthenticateParams struct {
	Email    string
	Password string
}
type AuthenticateResult struct {
	User PublicUser
}

// === List ===

// ListParams — вход List с пагинацией.
//
// Limit/Offset нормализуются в usecase: дефолт 20, максимум 100,
// отрицательный offset → 0.
type ListParams struct {
	Limit  int
	Offset int
}

// ListResult — выход List.
//
// Total — общее количество записей для пагинации.
// Limit/Offset — фактически применённые (после нормализации).
type ListResult struct {
	Users  []PublicUser
	Total  int64
	Limit  int
	Offset int
}

// ============================================================
// Исходящие порты
// ============================================================

// Repository — что usecase требует от хранилища.
//
// Реализуется:
//   - pgRepository (Postgres)
//   - cachedRepository (декоратор с Redis)
type Repository interface {
	Save(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context, limit, offset int) ([]User, error)
	Count(ctx context.Context) (int64, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// Hasher — хеширование и проверка пароля.
type Hasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}
