package token

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// FakeIssuer — для тестов. Детерминированные токены.
// Не использовать в проде.
type FakeIssuer struct{}

func NewFakeIssuer() *FakeIssuer { return &FakeIssuer{} }

var _ Issuer = (*FakeIssuer)(nil)

func (FakeIssuer) Issue(ctx context.Context, userID uuid.UUID) (Pair, error) {
	return Pair{
		AccessToken:  fmt.Sprintf("fake-access-%s", userID),
		RefreshToken: fmt.Sprintf("fake-refresh-%s", userID),
	}, nil
}

func (FakeIssuer) ValidateAccess(ctx context.Context, token string) (Claims, error) {
	return parseFake(token, TypeAccess)
}

func (FakeIssuer) ValidateRefresh(ctx context.Context, token string) (Claims, error) {
	return parseFake(token, TypeRefresh)
}

func parseFake(s string, expected TokenType) (Claims, error) {
	var (
		prefix string
		typ    TokenType
	)

	switch expected {
	case TypeAccess:
		prefix = "fake-access-"
		typ = TypeAccess
	case TypeRefresh:
		prefix = "fake-refresh-"
		typ = TypeRefresh
	default:
		return Claims{}, ErrInvalidToken
	}

	if len(s) <= len(prefix) || s[:len(prefix)] != prefix {
		return Claims{}, ErrInvalidToken
	}

	userID, err := uuid.Parse(s[len(prefix):])
	if err != nil {
		return Claims{}, ErrInvalidToken
	}

	return Claims{Subject: userID, Type: typ}, nil
}
