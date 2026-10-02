package users

import "context"

// GetByID возвращает пользователя по ID.
//
// Ошибки:
//   - ErrNotFound → 404 (маппится в errs.NotFound).
//   - Прочие      → 500.
func (u *usecase) GetByID(ctx context.Context, params GetByIDParams) (GetByIDResult, error) {
	user, err := u.repo.GetByID(ctx, params.ID)
	if err != nil {
		return GetByIDResult{}, mapDomainError(err).WithOp(opGetByID)
	}
	return GetByIDResult{User: user.Public()}, nil
}
