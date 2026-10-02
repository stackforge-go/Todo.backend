package users

import (
	"context"

	"github.com/google/uuid"
)

// ============================================================
// Входящий порт
// ============================================================

// Usecase — публичный контракт фичи users.
type Usecase interface {
	Create(ctx context.Context, params CreateParams) (CreateResult, error)
	GetByID(ctx context.Context, params GetByIDParams) (GetByIDResult, error)
	List(ctx context.Context, params ListParams) (ListResult, error)
}

// ============================================================
// Params / Result
// ============================================================

type CreateParams struct {
	Email         string
	Password      string
	FullName      string
	EmailVerified bool
}

type CreateResult struct {
	User PublicUser
}

type GetByIDParams struct {
	ID uuid.UUID
}
type GetByIDResult struct {
	User PublicUser
}

type ListParams struct {
	Limit  int
	Offset int
}
type ListResult struct {
	Users  []PublicUser
	Total  int64
	Limit  int
	Offset int
}

// ============================================================
// Исходящие порты
// ============================================================

type Repository interface {
	Save(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	List(ctx context.Context, limit, offset int) ([]User, error)
	Count(ctx context.Context) (int64, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type Hasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}
