package token

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// TokenType — тип токена.
type TokenType string

const (
	TypeAccess  TokenType = "access"
	TypeRefresh TokenType = "refresh"
)

// Claims — что извлекается из токена.
type Claims struct {
	Subject   uuid.UUID
	Type      TokenType
	IssuedAt  time.Time
	ExpiresAt time.Time
}

// Pair — пара токенов.
type Pair struct {
	AccessToken  string
	RefreshToken string
}

// Issuer — выдача и проверка токенов.
type Issuer interface {
	// Issue выдаёт пару токенов для пользователя.
	Issue(ctx context.Context, userID uuid.UUID) (Pair, error)

	// ValidateAccess проверяет access-токен.
	ValidateAccess(ctx context.Context, token string) (Claims, error)

	// ValidateRefresh проверяет refresh-токен.
	ValidateRefresh(ctx context.Context, token string) (Claims, error)
}
