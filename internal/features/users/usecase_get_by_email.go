package users

import "context"

// GetByEmail возвращает пользователя по email.
//
// НЕ используется для аутентификации — PasswordHash не покидает users.
// Для логина используй Authenticate.
func (u *usecase) GetByEmail(ctx context.Context, params GetByEmailParams) (GetByEmailResult, error) {
	user, err := u.repo.GetByEmail(ctx, params.Email)
	if err != nil {
		return GetByEmailResult{}, mapDomainError(err).WithOp(opGetByEmail)
	}
	return GetByEmailResult{User: user.Public()}, nil
}
