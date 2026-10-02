package auth

import (
	"context"

	"github.com/stackforge-go/Todo.backend/internal/features/users"
	"github.com/stackforge-go/Todo.backend/internal/infrastructure/errs"
)

// Register регистрирует нового пользователя.
//
// Шаги:
//  1. Создаёт пользователя через users.Usecase.Create с EmailVerified=false.
//  2. Возвращает PublicUser + сообщение о необходимости подтвердить email.
//
// Токенов НЕТ: пользователь должен сначала подтвердить email.
// Логин (выдача токенов) — отдельный эндпоинт.
//
// Отправка welcome/verification email — ответственность другого
// слоя (notifications consumer, реагирующий на событие user.registered).
//
// Ошибки users (ErrEmailTaken, ErrEmailInvalid) пробрасываются
// как errs.AppError через errs.Wrap — с добавлением op=auth.Register.
func (u *usecase) Register(ctx context.Context, params RegisterParams) (RegisterResult, error) {
	// 1. Создать пользователя. EmailVerified=false — нужно подтверждение.
	out, err := u.users.Create(ctx, users.CreateParams{
		Email:         params.Email,
		Password:      params.Password,
		FullName:      params.FullName,
		EmailVerified: false,
	})
	if err != nil {
		// err уже errs.AppError — Wrap вернёт тот же указатель,
		// WithOp добавит контекст auth.
		return RegisterResult{}, errs.Wrap(err).WithOp(opRegister)
	}

	// 2. Вернуть сообщение. Токены не выдаём — сначала верификация.
	return RegisterResult{
		User:    out.User,
		Message: RegisterMessage,
	}, nil
}
