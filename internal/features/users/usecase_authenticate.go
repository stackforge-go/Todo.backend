package users

import (
	"context"
)

// Authenticate проверяет email и пароль пользователя.
//
// Шаги:
//  1. Находит пользователя по email (через кэш).
//  2. Сравнивает хеш пароля через Hasher.Compare.
//
// Безопасность: при ЛЮБОЙ неудаче возвращает ErrInvalidCredentials.
// Клиент не может отличить «email не найден» от «пароль неверный» —
// защита от user enumeration.
//
// PasswordHash никогда не покидает users: наружу возвращается
// только PublicUser.
func (u *usecase) Authenticate(
	ctx context.Context,
	params AuthenticateParams,
) (AuthenticateResult, error) {
	// 1. Найти пользователя.
	user, err := u.repo.GetByEmail(ctx, params.Email)
	if err != nil {
		// Не раскрываем причину: ErrNotFound тоже → ErrInvalidCredentials.
		return AuthenticateResult{}, ErrInvalidCredentials
	}

	// 2. Проверить пароль.
	if err := u.hasher.Compare(user.PasswordHash, params.Password); err != nil {
		return AuthenticateResult{}, ErrInvalidCredentials
	}

	return AuthenticateResult{User: user.Public()}, nil
}
