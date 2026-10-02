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
	// maxEmailLen — максимальная длина email в символах.
	// Соответствует VARCHAR(254) в БД (стандарт RFC 5321).
	maxEmailLen = 254

	// maxFullNameLen — максимальная длина full_name в символах.
	// Соответствует VARCHAR(150) в БД.
	maxFullNameLen = 150
)

// ============================================================
// Ошибки домена
// ============================================================

// Доменные ошибки. Простые errors.New — без зависимости
// от infrastructure. Маппятся в errs.AppError на уровне usecase.
var (
	ErrEmailRequired    = errors.New("email is required")
	ErrEmailInvalid     = errors.New("email invalid")
	ErrFullNameTooLong  = errors.New("full_name too long")
	ErrPasswordRequired = errors.New("password hash is required")
	ErrEmailTaken       = errors.New("email already taken")
	ErrNotFound         = errors.New("user not found")

	// ErrInvalidCredentials — email не найден ИЛИ пароль неверный.
	// Одна ошибка на оба случая — защита от user enumeration:
	// клиент не может узнать, зарегистрирован ли email в системе.
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// ============================================================
// Доменная сущность
// ============================================================

// User — доменная сущность. Внутренняя, не покидает фичу.
//
// Содержит приватные поля (PasswordHash) и технические (Version).
// Наружу отдаётся PublicUser — без PasswordHash и Version.
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

// CreateUser — единственный способ создать нового пользователя.
//
// Задаёт умолчания:
//   - Version = 1             — первая версия агрегата
//   - CreatedAt = UpdatedAt = now
//   - Email нормализуется (lower + trim)
//   - FullName тримится
//
// Валидация вызывается внутри — возвращённый User всегда валиден.
// EmailVerified передаётся снаружи: auth → false, admin → true.
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

// Reconstitute — восстановление User из БД. Все поля явно.
//
// Используется только репозиторием. Никаких умолчаний и валидации —
// данные уже валидны в БД (при записи прошли Validate).
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
//
// Длина считается в СИМВОЛАХ (utf8.RuneCountInString), а не в байтах,
// чтобы совпадать с VARCHAR(N) в Postgres.
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
// Вызывается перед Update: Version используется для оптимистичной
// блокировки (ON CONFLICT ... WHERE version = ...).
func (u *User) Touch(now time.Time) {
	u.Version++
	u.UpdatedAt = now.UTC()
}

// ============================================================
// Публичное представление
// ============================================================

// PublicUser — публичное представление пользователя.
//
// Не содержит PasswordHash и Version — их нельзя показывать наружу.
// Возвращается из Usecase: другим фичам, транспорту.
//
// Value object: по значению, не по указателю.
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
