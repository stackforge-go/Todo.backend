package auth

import "errors"

// ============================================================
// Ошибки домена
// ============================================================

// Доменные ошибки. Простые errors.New — без зависимости
// от infrastructure. Маппятся в errs.AppError на уровне usecase.
var (
	// ErrInvalidCredentials — неверный email или пароль.
	// Не различаем "email не найден" и "пароль неверный" —
	// защита от user enumeration.
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrInvalidToken — токен невалиден, истёк или неверной подписи.
	ErrInvalidToken = errors.New("invalid token")

	// ErrEmailNotVerified — email не подтверждён.
	ErrEmailNotVerified = errors.New("email not verified")
)

// ============================================================
// Константы
// ============================================================

// RegisterMessage — сообщение клиенту после успешной регистрации.
//
// Клиент показывает его пользователю. Формулировка нейтральная —
// не раскрывает деталей, но объясняет, что делать дальше.
const RegisterMessage = "Registration successful. Please check your email to verify your account."
