package golangjwt

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/stackforge-go/Todo.backend/internal/infrastructure/token"
)

// signingMethod — HS256. Симметричная подпись.
//
// var, не const: SigningMethodHS256 — переменная в golang-jwt.
var signingMethod = jwt.SigningMethodHS256

// Issuer — реализация token.Issuer через github.com/golang-jwt/jwt/v5.
type Issuer struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
	issuer     string
}

func NewIssuer(cfg Config) *Issuer {
	return &Issuer{
		secret:     []byte(cfg.Secret),
		accessTTL:  cfg.AccessTTL,
		refreshTTL: cfg.RefreshTTL,
		issuer:     cfg.Issuer,
	}
}

var _ token.Issuer = (*Issuer)(nil)

// ============================================================
// Issue
// ============================================================

func (i *Issuer) Issue(ctx context.Context, userID uuid.UUID) (token.Pair, error) {
	access, err := i.sign(userID, token.TypeAccess, i.accessTTL)
	if err != nil {
		return token.Pair{}, fmt.Errorf("sign access token: %w", err)
	}

	refresh, err := i.sign(userID, token.TypeRefresh, i.refreshTTL)
	if err != nil {
		return token.Pair{}, fmt.Errorf("sign refresh token: %w", err)
	}

	return token.Pair{
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

// ============================================================
// Validate
// ============================================================

// ValidateAccess проверяет access-токен.
func (i *Issuer) ValidateAccess(ctx context.Context, tokenStr string) (token.Claims, error) {
	return i.validate(tokenStr, token.TypeAccess)
}

// ValidateRefresh проверяет refresh-токен.
func (i *Issuer) ValidateRefresh(ctx context.Context, tokenStr string) (token.Claims, error) {
	return i.validate(tokenStr, token.TypeRefresh)
}

// ============================================================
// Внутреннее
// ============================================================

// claims — JWT-специфичные claims. Маппятся в token.Claims.
type claims struct {
	Type string `json:"typ"`
	jwt.RegisteredClaims
}

func (i *Issuer) sign(userID uuid.UUID, typ token.TokenType, ttl time.Duration) (string, error) {
	now := time.Now()

	c := claims{
		Type: string(typ),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    i.issuer,
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	t := jwt.NewWithClaims(signingMethod, c)

	signed, err := t.SignedString(i.secret)
	if err != nil {
		return "", fmt.Errorf("sign: %w", err)
	}

	return signed, nil
}

func (i *Issuer) validate(tokenStr string, expected token.TokenType) (token.Claims, error) {
	parsed := &claims{}

	t, err := jwt.ParseWithClaims(
		tokenStr,
		parsed,
		func(t *jwt.Token) (any, error) {
			// Защита от alg=none и подмены алгоритма.
			if t.Method != signingMethod {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return i.secret, nil
		},
		jwt.WithIssuer(i.issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return token.Claims{}, token.ErrInvalidToken
	}
	if !t.Valid {
		return token.Claims{}, token.ErrInvalidToken
	}

	// Проверка типа токена.
	if token.TokenType(parsed.Type) != expected {
		return token.Claims{}, token.ErrInvalidToken
	}

	userID, err := uuid.Parse(parsed.Subject)
	if err != nil {
		return token.Claims{}, token.ErrInvalidToken
	}

	return token.Claims{
		Subject:   userID,
		Type:      token.TokenType(parsed.Type),
		IssuedAt:  parsed.IssuedAt.Time,
		ExpiresAt: parsed.ExpiresAt.Time,
	}, nil
}
