package users

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/stackforge-go/Todo.backend/internal/infrastructure/errs"
)

// Create создаёт нового пользователя.
//
// Шаги:
//  1. Хеширует сырой пароль через Hasher.
//  2. Генерирует UUID v7 (time-ordered).
//  3. Создаёт доменную сущность через CreateUser (валидация внутри).
//  4. Сохраняет через Repository.Save.
//  5. Возвращает PublicUser — без PasswordHash.
//
// Уникальность email обеспечивается UNIQUE-индексом в БД:
// pgRepository.Save маппит unique violation → ErrEmailTaken.
// Отдельного SELECT перед INSERT нет — гонка всё равно возможна,
// UNIQUE — единственный источник истины.
func (s *usecase) Create(ctx context.Context, params CreateParams) (CreateResult, error) {
	// 1. Хеш пароля. Сырой пароль дальше не используется.
	hash, err := s.hasher.Hash(params.Password)
	if err != nil {
		return CreateResult{}, errs.Internal.
			WithMessage("hash password").
			WithOp(opCreate).
			Wrap(err)
	}

	// 2. ID. UUID v7 — time-ordered, хорошо для индексов.
	id, err := uuid.NewV7()
	if err != nil {
		return CreateResult{}, errs.Internal.
			WithMessage("generate id").
			WithOp(opCreate).
			Wrap(err)
	}

	// 3. Домен. CreateUser задаёт Version=1, CreatedAt/UpdatedAt, валидирует.
	u, err := CreateUser(
		id,
		params.Email,
		params.EmailVerified,
		hash,
		params.FullName,
		time.Now().UTC(),
	)
	if err != nil {
		return CreateResult{}, mapDomainError(err).WithOp(opCreate)
	}

	// 4. Save. ErrEmailTaken придёт из pgRepository при unique violation.
	if err := s.repo.Save(ctx, u); err != nil {
		return CreateResult{}, mapDomainError(err).WithOp(opCreate)
	}

	// 5. Наружу — PublicUser.
	return CreateResult{User: u.Public()}, nil
}
