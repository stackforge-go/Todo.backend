package auth

import (
	"context"

	"github.com/stackforge-go/Todo.backend/internal/features/users"
	"github.com/stackforge-go/Todo.backend/internal/infrastructure/errs"
)

// Refresh обновляет access-токен по refresh-токену.
//
// Шаги:
//  1. Валидирует refresh-токен через TokenIssuer.ValidateRefresh.
//     — Неверная подпись / срок / тип → ErrInvalidToken (401).
//  2. Получает пользователя по userID из токена (с кэшем).
//     — Пользователь удалён → ErrNotFound (404).
//  3. Выдаёт новую пару токенов.
//
// Тип токена проверяется внутри Issuer.ValidateRefresh:
// access-токен сюда не пройдёт.
func (u *usecase) Refresh(ctx context.Context, params RefreshParams) (RefreshResult, error) {
	// 1. Валидация refresh-токена.
	claims, err := u.tokens.ValidateRefresh(ctx, params.RefreshToken)
	if err != nil {
		return RefreshResult{}, mapDomainError(ErrInvalidToken).WithOp(opRefresh)
	}

	// 2. Получаем пользователя. Может быть удалён.
	out, err := u.users.GetByID(ctx, users.GetByIDParams{ID: claims.Subject})
	if err != nil {
		return RefreshResult{}, errs.Wrap(err).WithOp(opRefresh)
	}

	// 3. Новая пара токенов для того же пользователя.
	tokens, err := u.tokens.Issue(ctx, claims.Subject)
	if err != nil {
		return RefreshResult{}, errs.Internal.
			WithMessage("issue tokens").
			WithOp(opRefresh).
			Wrap(err)
	}

	return RefreshResult{
		User:   out.User,
		Tokens: tokens,
	}, nil
}
