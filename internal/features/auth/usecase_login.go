package auth

import (
	"context"

	"github.com/stackforge-go/Todo.backend/internal/features/users"
	"github.com/stackforge-go/Todo.backend/internal/infrastructure/errs"
)

// Login аутентифицирует пользователя по email и паролю.
//
// Шаги:
//  1. users.Authenticate проверяет email + пароль.
//     — Найден и пароль верный → PublicUser.
//     — Не найден ИЛИ пароль неверный → ErrInvalidCredentials (401).
//  2. Выдаёт пару токенов.
//
// auth НЕ работает с PasswordHash: вся работа с хешем — внутри users.
// Если email не найден — возвращаем ту же ошибку, что при неверном
// пароле (защита от user enumeration — внутри users.Authenticate).
func (u *usecase) Login(ctx context.Context, params LoginParams) (LoginResult, error) {
	// 1. Аутентификация.
	out, err := u.users.Authenticate(ctx, users.AuthenticateParams{
		Email:    params.Email,
		Password: params.Password,
	})
	if err != nil {
		// err уже errs.AppError (Unauthorized с "invalid credentials").
		return LoginResult{}, errs.Wrap(err).WithOp(opLogin)
	}

	// 2. Токены.
	tokens, err := u.tokens.Issue(ctx, out.User.ID)
	if err != nil {
		return LoginResult{}, errs.Internal.
			WithMessage("issue tokens").
			WithOp(opLogin).
			Wrap(err)
	}

	return LoginResult{
		User:   out.User,
		Tokens: tokens,
	}, nil
}
