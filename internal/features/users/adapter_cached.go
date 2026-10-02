package users

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/stackforge-go/Todo.backend/internal/infrastructure/redis"
)

// cachedRepository — декоратор Repository с кэшем через Redis.
//
// Strategy: cache-aside.
//   - GetByID: GET  user:<id>
//   - List:    HGET users:all <limit>:<offset>
//   - Count:   GET  users:count
//   - Save:    SET user:<id> + DEL users:all, users:count
//   - Delete:  DEL user:<id> + DEL users:all, users:count
//
// Best-effort: ошибки кэша не прерывают запрос.
type cachedRepository struct {
	pool redis.Pool
	repo Repository
}

func NewCachedRepository(
	pool redis.Pool,
	repo Repository,
) *cachedRepository {
	return &cachedRepository{
		pool: pool,
		repo: repo,
	}
}

var _ Repository = (*cachedRepository)(nil)

// ============================================================
// Save
// ============================================================

func (r *cachedRepository) Save(ctx context.Context, u *User) error {
	if err := r.repo.Save(ctx, u); err != nil {
		return err
	}

	r.cacheUser(ctx, u)
	r.invalidateListsAndCount(ctx)

	return nil
}

// ============================================================
// GetByID
// ============================================================

func (r *cachedRepository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	if u, ok := r.getUserFromCache(ctx, id); ok {
		return u, nil
	}

	u, err := r.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	r.cacheUser(ctx, u)
	return u, nil
}

// ============================================================
// List
// ============================================================

func (r *cachedRepository) List(ctx context.Context, limit, offset int) ([]User, error) {
	key := usersListKey()
	field := usersListField(limit, offset)

	if raw, err := r.pool.HGet(ctx, key, field).Bytes(); err == nil {
		var users []User
		if err := json.Unmarshal(raw, &users); err == nil {
			return users, nil
		}
	}

	users, err := r.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	if raw, err := json.Marshal(users); err == nil {
		_ = r.pool.HSet(ctx, key, field, raw).Err()
	}

	return users, nil
}

// ============================================================
// Count
// ============================================================

func (r *cachedRepository) Count(ctx context.Context) (int64, error) {
	key := usersCountKey()

	if raw, err := r.pool.Get(ctx, key).Bytes(); err == nil {
		var total int64
		if err := json.Unmarshal(raw, &total); err == nil {
			return total, nil
		}
	}

	total, err := r.repo.Count(ctx)
	if err != nil {
		return 0, err
	}

	if raw, err := json.Marshal(total); err == nil {
		_ = r.pool.Set(ctx, key, raw, r.pool.TTL()).Err()
	}

	return total, nil
}

// ============================================================
// Delete
// ============================================================

func (r *cachedRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.repo.Delete(ctx, id); err != nil {
		return err
	}

	_ = r.pool.Del(ctx, userKeyByID(id)).Err()
	r.invalidateListsAndCount(ctx)

	return nil
}

// ============================================================
// Внутреннее
// ============================================================

func (r *cachedRepository) getUserFromCache(ctx context.Context, id uuid.UUID) (*User, bool) {
	raw, err := r.pool.Get(ctx, userKeyByID(id)).Bytes()
	if err != nil {
		return nil, false
	}

	var u User
	if err := json.Unmarshal(raw, &u); err != nil {
		return nil, false
	}

	return &u, true
}

func (r *cachedRepository) cacheUser(ctx context.Context, u *User) {
	raw, err := json.Marshal(u)
	if err != nil {
		return
	}

	_ = r.pool.Set(ctx, userKeyByID(u.ID), raw, r.pool.TTL()).Err()
}

func (r *cachedRepository) invalidateListsAndCount(ctx context.Context) {
	_ = r.pool.Del(ctx, usersListKey(), usersCountKey()).Err()
}

// ============================================================
// Ключи и поля
// ============================================================

func userKeyByID(id uuid.UUID) string {
	return fmt.Sprintf("user:%s", id)
}

func usersListKey() string {
	return "users:all"
}

func usersCountKey() string {
	return "users:count"
}

func usersListField(limit, offset int) string {
	return fmt.Sprintf("%d:%d", limit, offset)
}
