package users

import "context"

func (u *usecase) GetByID(ctx context.Context, params GetByIDParams) (GetByIDResult, error) {
	user, err := u.repo.GetByID(ctx, params.ID)
	if err != nil {
		return GetByIDResult{}, mapDomainError(err).WithOp(opGetByID)
	}
	return GetByIDResult{User: user.Public()}, nil
}
