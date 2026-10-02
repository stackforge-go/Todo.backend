package users

import (
	"context"
	"errors"
	"sync"

	"github.com/stackforge-go/Todo.backend/internal/infrastructure/errs"
)

const (
	defaultLimit = 20
	maxLimit     = 100

	opCreate  = "users.Create"
	opGetByID = "users.GetByID"
	opList    = "users.List"
)

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

var _ Usecase = (*usecase)(nil)

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

	default:
		return errs.Internal.WithMessage("internal error").Wrap(err)
	}
}

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
