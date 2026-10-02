package token

import "errors"

var (
	// ErrInvalidToken — токен невалиден (подпись, срок, тип).
	ErrInvalidToken = errors.New("invalid token")
)
