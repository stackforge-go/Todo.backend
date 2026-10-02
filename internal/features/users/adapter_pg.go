package users

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/stackforge-go/Todo.backend/internal/infrastructure/postgres"
)

// ============================================================
// SQL
// ============================================================

const (
	// selectColumns — общий набор колонок для всех SELECT.
	// Меняется здесь — меняется везде.
	selectColumns = `
		id, version, created_at, updated_at,
		email, email_verified, password_hash, full_name
	`

	// selectByID — одна строка по id.
	selectByID = `SELECT ` + selectColumns + ` FROM todo.users WHERE id = $1`

	// selectByID — одна строка по id.
	selectByEmail = `SELECT ` + selectColumns + ` FROM todo.users WHERE email = $1`

	// selectList — страница пользователей.
	// Стабильная сортировка: created_at DESC, id DESC
	// (id — тайбрейкер при одинаковых created_at).
	selectList = `SELECT ` + selectColumns + `
		FROM todo.users
		ORDER BY created_at DESC, id DESC
		LIMIT $1 OFFSET $2`

	// selectCount — общее количество для пагинации.
	selectCount = `SELECT COUNT(*) FROM todo.users`
)

// ============================================================
// Repository
// ============================================================

// pgRepository — реализация Repository через Postgres (pgx).
type pgRepository struct {
	pool postgres.Pool
}

func NewPgRepository(pool postgres.Pool) *pgRepository {
	return &pgRepository{pool: pool}
}

var _ Repository = (*pgRepository)(nil)

// ============================================================
// Save
// ============================================================

// Save делает UPSERT по id.
//
// При конфликте UNIQUE (email) возвращает ErrEmailTaken —
// это единственный источник истины для уникальности email.
func (r *pgRepository) Save(ctx context.Context, u *User) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	const q = `
		INSERT INTO todo.users (
			id, version, created_at, updated_at,
			email, email_verified, password_hash, full_name
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			version        = EXCLUDED.version,
			updated_at     = EXCLUDED.updated_at,
			email          = EXCLUDED.email,
			email_verified = EXCLUDED.email_verified,
			password_hash  = EXCLUDED.password_hash,
			full_name      = EXCLUDED.full_name
	`

	_, err := r.pool.Exec(ctx, q,
		u.ID,
		u.Version,
		u.CreatedAt,
		u.UpdatedAt,
		u.Email,
		u.EmailVerified,
		u.PasswordHash,
		nullableString(u.FullName),
	)
	if errors.Is(err, postgres.ErrViolatesUniqueKey) {
		return ErrEmailTaken
	}
	return err
}

// ============================================================
// GetByID
// ============================================================

// GetByID возвращает пользователя по id.
//
// ErrNotFound — если строки нет (маппится из postgres.ErrNoRows).
func (r *pgRepository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	return r.scanOne(ctx, selectByID, id)
}

// ============================================================
// GetByEmail
// ============================================================

// GetByEmail возвращает пользователя по email.
//
// ErrNotFound — если строки нет (маппится из postgres.ErrNoRows).
func (r *pgRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	return r.scanOne(ctx, selectByEmail, email)
}

// ============================================================
// List
// ============================================================

// List возвращает страницу пользователей.
//
// Лимит/офсет уже нормализованы в usecase.
func (r *pgRepository) List(ctx context.Context, limit, offset int) ([]User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	return r.scanMany(ctx, selectList, limit, offset)
}

// ============================================================
// Count
// ============================================================

// Count возвращает общее количество пользователей.
// Используется в usecase для пагинации.
func (r *pgRepository) Count(ctx context.Context) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	var total int64
	if err := r.pool.QueryRow(ctx, selectCount).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

// ============================================================
// Delete
// ============================================================

// Delete удаляет пользователя по id.
//
// Идемпотентен: удаление несуществующего — не ошибка.
func (r *pgRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	const q = `DELETE FROM todo.users WHERE id = $1`

	_, err := r.pool.Exec(ctx, q, id)
	return err
}

// ============================================================
// scan helpers
// ============================================================

// scanOne читает одну строку. Возвращает *User, потому что
// nil = «не найдено».
func (r *pgRepository) scanOne(ctx context.Context, q string, args ...any) (*User, error) {
	u, err := scanUser(r.pool.QueryRow(ctx, q, args...))
	if errors.Is(err, postgres.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// scanMany читает список строк. Возвращает []User.
// make([]User, 0) — не-nil срез, чтобы JSON был [] а не null.
func (r *pgRepository) scanMany(ctx context.Context, q string, args ...any) ([]User, error) {
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]User, 0)
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// scanUser — общий хелпер: сканирует строку в User (значение).
//
// Работает и для QueryRow, и для Rows — у обоих есть Scan.
// full_name в БД nullable → сканируем в *string.
func scanUser(row interface {
	Scan(dest ...any) error
}) (User, error) {
	var (
		u        User
		fullName *string
	)

	err := row.Scan(
		&u.ID,
		&u.Version,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.Email,
		&u.EmailVerified,
		&u.PasswordHash,
		&fullName,
	)
	if err != nil {
		return User{}, err
	}

	if fullName != nil {
		u.FullName = *fullName
	}

	return *Reconstitute(
		u.ID,
		u.Version,
		u.CreatedAt,
		u.UpdatedAt,
		u.Email,
		u.EmailVerified,
		u.PasswordHash,
		u.FullName,
	), nil
}

// nullableString — "" → nil, иначе &s.
// Для nullable-полей в БД (full_name).
func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
