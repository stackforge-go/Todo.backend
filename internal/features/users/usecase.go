package users

import (
	"context"
	"errors"
	"sync"

	"github.com/stackforge-go/Todo.backend/internal/infrastructure/errs"
)

// ============================================================
// Константы
// ============================================================

const (
	// defaultLimit — размер страницы по умолчанию.
	defaultLimit = 20

	// maxLimit — максимум записей за один запрос.
	// Защита от ?limit=1000000.
	maxLimit = 100

	// op* — операции для логов. Используются в WithOp,
	// чтобы в логах был op="users.Create", а не голый error.
	opCreate       = "users.Create"
	opGetByID      = "users.GetByID"
	opGetByEmail   = "users.GetByEmail"
	opList         = "users.List"
	opAuthenticate = "users.Authenticate"
)

// ============================================================
// Реализация
// ============================================================

// usecase — реализация Usecase. Приватная.
//
// Зависимости — только интерфейсы (Repository, Hasher).
// Конкретные реализации (pgRepository, BcryptHasher) подставляются в main.go.
type usecase struct {
	repo   Repository
	hasher Hasher
}

func NewUsecase(repo Repository, hasher Hasher) *usecase {
	return &usecase{
		repo:   repo,
		hasher: hasher,
	}
}

// Compile-time проверка: *usecase реализует Usecase.
var _ Usecase = (*usecase)(nil)

// ============================================================
// Маппинг доменных ошибок
// ============================================================

// mapDomainError конвертирует доменные ошибки в errs.AppError.
//
// Домен не знает про errs — это ответственность usecase.
// Транспорт (HTTP, gRPC) работает только с errs.AppError.
func mapDomainError(err error) *errs.AppError {
	switch {
	case errors.Is(err, ErrEmailRequired),
		errors.Is(err, ErrEmailInvalid),
		errors.Is(err, ErrFullNameTooLong),
		errors.Is(err, ErrPasswordRequired):
		return errs.InvalidArgument.WithMessage(err.Error()).Wrap(err)

	case errors.Is(err, ErrEmailTaken):
		return errs.AlreadyExists.WithMessage(err.Error()).Wrap(err)

	case errors.Is(err, ErrNotFound):
		return errs.NotFound.WithMessage(err.Error()).Wrap(err)

	case errors.Is(err, ErrInvalidCredentials):
		return errs.Unauthorized.WithMessage("invalid credentials").Wrap(err)

	default:
		return errs.Internal.WithMessage("internal error").Wrap(err)
	}
}

// ============================================================
// Хелперы
// ============================================================

// normalizePagination приводит limit/offset к допустимым значениям:
//   - limit <= 0       → defaultLimit
//   - limit > maxLimit → maxLimit
//   - offset < 0       → 0
func normalizePagination(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

// listParallel параллельно вызывает repo.List и repo.Count.
//
// Каждая горутина пишет в СВОЮ переменную — гонки нет.
// wg.Wait() — memory barrier: после него все записи видны.
// Обе ошибки проверяются отдельно.
func (u *usecase) listParallel(ctx context.Context, limit, offset int) ([]User, int64, error) {
	var (
		users    []User
		total    int64
		usersErr error
		totalErr error
	)

	var wg sync.WaitGroup
	wg.Go(func() {
		users, usersErr = u.repo.List(ctx, limit, offset)
	})
	wg.Go(func() {
		total, totalErr = u.repo.Count(ctx)
	})
	wg.Wait()

	if usersErr != nil {
		return nil, 0, usersErr
	}
	if totalErr != nil {
		return nil, 0, totalErr
	}
	return users, total, nil
}
