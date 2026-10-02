package users

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// ============================================================
// Константы
// ============================================================

const (
	maxEmailLen    = 254
	maxFullNameLen = 150
)

// ============================================================
// Ошибки домена
// ============================================================

var (
	ErrEmailRequired    = errors.New("email is required")
	ErrEmailInvalid     = errors.New("email invalid")
	ErrFullNameTooLong  = errors.New("full_name too long")
	ErrPasswordRequired = errors.New("password hash is required")
	ErrEmailTaken       = errors.New("email already taken")
	ErrNotFound         = errors.New("user not found")
)

// ============================================================
// Доменная сущность
// ============================================================

// User — доменная сущность. Внутренняя, не покидает фичу.
//
// Содержит приватные поля (PasswordHash), технические (Version).
// Наружу отдаётся PublicUser.
type User struct {
	ID            uuid.UUID
	Version       int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Email         string
	PasswordHash  string
	FullName      string
	EmailVerified bool
}

// CreateUser — создание нового пользователя.
//
// Задаёт умолчания:
//   - Version = 1
//   - CreatedAt = UpdatedAt = now
//
// Снаружи передаётся только то, что реально варьируется.
func CreateUser(
	id uuid.UUID,
	email string,
	emailVerified bool,
	passwordHash string,
	fullName string,
	now time.Time,
) (*User, error) {
	u := &User{
		ID:            id,
		Version:       1,
		CreatedAt:     now.UTC(),
		UpdatedAt:     now.UTC(),
		Email:         strings.ToLower(strings.TrimSpace(email)),
		EmailVerified: emailVerified,
		PasswordHash:  passwordHash,
		FullName:      strings.TrimSpace(fullName),
	}

	if err := u.Validate(); err != nil {
		return nil, err
	}

	return u, nil
}

// Reconstitute — восстановление из БД. Все поля явно.
//
// Используется только репозиторием. Никаких умолчаний —
// данные уже валидны в БД.
func Reconstitute(
	id uuid.UUID,
	version int64,
	createdAt, updatedAt time.Time,
	email string,
	emailVerified bool,
	passwordHash string,
	fullName string,
) *User {
	return &User{
		ID:            id,
		Version:       version,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
		Email:         email,
		EmailVerified: emailVerified,
		PasswordHash:  passwordHash,
		FullName:      fullName,
	}
}

// Validate проверяет инварианты сущности.
func (u *User) Validate() error {
	if u.Email == "" {
		return ErrEmailRequired
	}
	if utf8.RuneCountInString(u.Email) > maxEmailLen || !strings.Contains(u.Email, "@") {
		return ErrEmailInvalid
	}
	if utf8.RuneCountInString(u.FullName) > maxFullNameLen {
		return ErrFullNameTooLong
	}
	if u.PasswordHash == "" {
		return ErrPasswordRequired
	}
	return nil
}

// Touch обновляет UpdatedAt и инкрементит Version.
//
// Вызывается перед Update, чтобы оптимистичная блокировка работала.
func (u *User) Touch(now time.Time) {
	u.Version++
	u.UpdatedAt = now.UTC()
}

// ============================================================
// Публичное представление
// ============================================================

// PublicUser — публичное представление пользователя.
//
// Не содержит приватных полей (PasswordHash) и технических (Version).
// Возвращается из Service наружу: другим фичам, транспорту.
//
// Это value object — возвращается по значению, не по указателю.
type PublicUser struct {
	ID            uuid.UUID
	Email         string
	FullName      string
	EmailVerified bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Public конвертирует доменную сущность в публичное представление.
func (u *User) Public() PublicUser {
	return PublicUser{
		ID:            u.ID,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
		FullName:      u.FullName,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
}
