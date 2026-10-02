package users

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/stackforge-go/Todo.backend/internal/infrastructure/errs"
)

// ============================================================
// Константы
// ============================================================

const (
	defaultLimit = 20
	maxLimit     = 100
)

// ============================================================
// Публичный контракт
// ============================================================

type Service interface {
	Create(ctx context.Context, in CreateParams) (CreateResult, error)
	FindByID(ctx context.Context, in FindByIDParams) (FindByIDResult, error)
	List(ctx context.Context, in ListParams) (ListResult, error)
}

// ============================================================
// Зависимости
// ============================================================

type Repository interface {
	Save(ctx context.Context, u *User) error
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	List(ctx context.Context, limit int, offset int) ([]User, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
}

type Hasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}

// ============================================================
// Реализация
// ============================================================

type service struct {
	repo   Repository
	hasher Hasher
}

func NewService(
	repo Repository,
) *service {
	return &service{
		repo: repo,
	}
}

var _ Service = (*service)(nil)

// ============================================================
// Маппинг доменных ошибок
// ============================================================

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

// ============================================================
// Создание пользователя
// ============================================================

type CreateParams struct {
	Email         string
	Password      string
	FullName      string
	EmailVerified bool
}

type CreateResult struct {
	User PublicUser
}

func (s *service) Create(ctx context.Context, params CreateParams) (CreateResult, error) {
	hash, err := s.hasher.Hash(params.Password)
	if err != nil {
		return CreateResult{}, errs.Internal.WithMessage("hash password").WithOp("users.Create").Wrap(err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return CreateResult{}, errs.Internal.WithMessage("generate id").WithOp("users.Create").Wrap(err)
	}

	u, err := CreateUser(id, params.Email, params.EmailVerified, hash, params.FullName, time.Now().UTC())
	if err != nil {
		return CreateResult{}, mapDomainError(err).WithOp("users.Create")
	}

	if err := s.repo.Save(ctx, u); err != nil {
		return CreateResult{}, mapDomainError(err).WithOp("users.Create")
	}

	return CreateResult{User: u.Public()}, nil
}

// ============================================================
// Получение пользователя по ID
// ============================================================

type FindByIDParams struct {
	ID uuid.UUID
}
type FindByIDResult struct {
	User PublicUser
}

func (s *service) FindByID(ctx context.Context, params FindByIDParams) (FindByIDResult, error) {
	u, err := s.repo.FindByID(ctx, params.ID)
	if err != nil {
		return FindByIDResult{}, mapDomainError(err).WithOp("users.FindByID")
	}
	return FindByIDResult{User: u.Public()}, nil
}

// ============================================================
// Получение списка пользователя с пагинацией LIMIT и OFFSET
// ============================================================

type ListParams struct {
	Limit  int
	Offset int
}
type ListResult struct {
	Users  []PublicUser
	Total  int64
	Limit  int
	Offset int
}

func (s *service) List(ctx context.Context, params ListParams) (ListResult, error) {
	if params.Limit <= 0 {
		params.Limit = defaultLimit
	}
	if params.Limit > maxLimit {
		params.Limit = maxLimit
	}
	if params.Offset < 0 {
		params.Offset = 0
	}

	var (
		users    []User
		total    int64
		usersErr error
		totalErr error
	)

	var wg sync.WaitGroup
	wg.Go(func() {
		users, usersErr = s.repo.List(ctx, params.Limit, params.Offset)
	})
	wg.Go(func() {
		total, totalErr = s.repo.Count(ctx)
	})
	wg.Wait()

	if usersErr != nil {
		return ListResult{}, mapDomainError(usersErr).WithOp("users.List")
	}

	if totalErr != nil {
		return ListResult{}, mapDomainError(totalErr).WithOp("users.List")
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
