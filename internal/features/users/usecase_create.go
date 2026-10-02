package users

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/stackforge-go/Todo.backend/internal/infrastructure/errs"
)

func (s *usecase) Create(ctx context.Context, params CreateParams) (CreateResult, error) {
	hash, err := s.hasher.Hash(params.Password)
	if err != nil {
		return CreateResult{}, errs.Internal.WithMessage("hash password").WithOp(opCreate).Wrap(err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return CreateResult{}, errs.Internal.WithMessage("generate id").WithOp(opCreate).Wrap(err)
	}

	u, err := CreateUser(id, params.Email, params.EmailVerified, hash, params.FullName, time.Now().UTC())
	if err != nil {
		return CreateResult{}, mapDomainError(err).WithOp(opCreate)
	}

	if err := s.repo.Save(ctx, u); err != nil {
		return CreateResult{}, mapDomainError(err).WithOp(opCreate)
	}

	return CreateResult{User: u.Public()}, nil
}
