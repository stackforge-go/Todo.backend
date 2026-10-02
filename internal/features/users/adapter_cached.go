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
// Стратегия: cache-aside.
//
//	GetByID:    GET  user:<id>
//	GetByEmail: GET  user:email:<email>
//	List:       HGET users:all <limit>:<offset>
//	Count:      GET  users:count
//	Save:       SET  user:<id> + user:email:<email>
//	           DEL  users:all, users:count
//	Delete:     DEL  user:<id>, user:email:<email>
//	           DEL  users:all, users:count
//
// Best-effort: ошибки кэша логируются и не прерывают запрос.
// При недоступности Redis данные возвращаются из БД.
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

// Save пишет в БД, затем:
//   - кэширует user по ID и email (write-through)
//   - инвалидирует списки и Count — они устарели
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

// GetByID — cache-aside:
//
//  1. Пробуем кэш → hit → возврат.
//  2. Miss → читаем из БД → кэшируем → возврат.
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
// GetByEmail
// ============================================================

// GetByEmail — cache-aside:
//
//  1. Пробуем кэш по user:email:<email> → hit → возврат.
//  2. Miss → читаем из БД → кэшируем по обоим ключам → возврат.
func (r *cachedRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	if u, ok := r.getUserFromCacheByEmail(ctx, email); ok {
		return u, nil
	}

	u, err := r.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	r.cacheUser(ctx, u) // кладёт по обоим ключам
	return u, nil
}

// ============================================================
// List
// ============================================================

// List — cache-aside через Redis hash:
//
//	key   = "users:all"
//	field = "<limit>:<offset>"
//
// Все варианты пагинации хранятся под одним ключом — инвалидация
// одним DEL по ключу hash.
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

// Count — отдельный ключ "users:count".
// Инвалидируется при Save/Delete.
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

// Delete удаляет из БД, инвалидирует кэш:
//   - DEL user:<id>, user:email:<email>
//   - DEL users:all, users:count
//
// Email нужен для инвалидации user:email:<email>.
// Достаём пользователя до удаления (best-effort).
func (r *cachedRepository) Delete(ctx context.Context, id uuid.UUID) error {
	u, _ := r.repo.GetByID(ctx, id)

	if err := r.repo.Delete(ctx, id); err != nil {
		return err
	}

	keys := []string{userKeyByID(id)}
	if u != nil {
		keys = append(keys, userKeyByEmail(u.Email))
	}
	_ = r.pool.Del(ctx, keys...).Err()

	r.invalidateListsAndCount(ctx)

	return nil
}

// ============================================================
// Внутреннее
// ============================================================

// getUserFromCache читает User из кэша по id.
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

// getUserFromCacheByEmail читает User из кэша по email.
func (r *cachedRepository) getUserFromCacheByEmail(ctx context.Context, email string) (*User, bool) {
	raw, err := r.pool.Get(ctx, userKeyByEmail(email)).Bytes()
	if err != nil {
		return nil, false
	}

	var u User
	if err := json.Unmarshal(raw, &u); err != nil {
		return nil, false
	}

	return &u, true
}

// cacheUser сериализует User и кладёт в кэш по двум ключам:
//   - user:<id>
//   - user:email:<email>
func (r *cachedRepository) cacheUser(ctx context.Context, u *User) {
	raw, err := json.Marshal(u)
	if err != nil {
		return
	}

	_ = r.pool.Set(ctx, userKeyByID(u.ID), raw, r.pool.TTL()).Err()
	_ = r.pool.Set(ctx, userKeyByEmail(u.Email), raw, r.pool.TTL()).Err()
}

// invalidateListsAndCount сбрасывает списки и Count одним DEL.
func (r *cachedRepository) invalidateListsAndCount(ctx context.Context) {
	_ = r.pool.Del(ctx, usersListKey(), usersCountKey()).Err()
}

// ============================================================
// Ключи и поля
// ============================================================

func userKeyByID(id uuid.UUID) string {
	return fmt.Sprintf("user:%s", id)
}

func userKeyByEmail(email string) string {
	return fmt.Sprintf("user:email:%s", email)
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
