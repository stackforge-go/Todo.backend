package users

import (
	"context"
)

func (u *usecase) List(ctx context.Context, params ListParams) (ListResult, error) {
	limit, offset := normalizePagination(params.Limit, params.Offset)

	users, total, err := u.listParallel(ctx, limit, offset)
	if err != nil {
		return ListResult{}, mapDomainError(err).WithOp(opList)
	}

	result := make([]PublicUser, 0, len(users))
	for _, u := range users {
		result = append(result, u.Public())
	}

	return ListResult{
		Users:  result,
		Total:  total,
		Limit:  params.Limit,
		Offset: params.Offset,
	}, nil
}
