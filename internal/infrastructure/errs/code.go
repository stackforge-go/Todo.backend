package errs

type Code string

const (
	CodeUnknown         Code = "UNKNOWN"
	CodeInternal        Code = "INTERNAL"
	CodeInvalidArgument Code = "INVALID_ARGUMENT"
	CodeNotFound        Code = "NOT_FOUND"
	CodeAlreadyExists   Code = "ALREADY_EXISTS"
	CodeConflict        Code = "CONFLICT"
	CodeUnauthorized    Code = "UNAUTHORIZED"
	CodeForbidden       Code = "FORBIDDEN"
	CodeUnavailable     Code = "UNAVAILABLE"
	CodeTimeout         Code = "TIMEOUT"
	CodeRateLimited     Code = "RATE_LIMITED"
)
