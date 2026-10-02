package users

import "context"

// List возвращает страницу пользователей + общее количество.
//
// Шаги:
//  1. Нормализация limit/offset (дефолт 20, max 100, offset >= 0).
//  2. Параллельно: repo.List (страница) + repo.Count (total).
//  3. Маппинг []User → []PublicUser.
//
// В ответе — фактически применённые Limit/Offset, а не сырые
// из запроса: клиент видит реальные значения после нормализации.
func (u *usecase) List(ctx context.Context, params ListParams) (ListResult, error) {
	limit, offset := normalizePagination(params.Limit, params.Offset)

	users, total, err := u.listParallel(ctx, limit, offset)
	if err != nil {
		return ListResult{}, mapDomainError(err).WithOp(opList)
	}

	// Итерация по индексу: Public() с pointer receiver — копий нет.
	result := make([]PublicUser, 0, len(users))
	for i := range users {
		result = append(result, users[i].Public())
	}

	return ListResult{
		Users:  result,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}
