package domain

import "errors"

var (
	ErrNotFound             = errors.New("resource not found")
	ErrUnauthorized         = errors.New("unauthorized access")
	ErrForbidden            = errors.New("forbidden operation")
	ErrBadRequest           = errors.New("invalid request payload")
	ErrConflict             = errors.New("resource already exists")
	ErrInternal             = errors.New("internal server error")
	ErrValidation           = errors.New("validation failed")
	ErrInvalidToken         = errors.New("invalid or expired token")
	ErrInvalidCredentials   = errors.New("kredensial login tidak valid")
	ErrUserInactive         = errors.New("akun pengguna dinonaktifkan")
	ErrEmailAlreadyInUse    = errors.New("alamat email sudah terdaftar")
	ErrEmailInvalid         = errors.New("format alamat email tidak valid")
	ErrPasswordTooShort     = errors.New("panjang password minimal 8 karakter")
	ErrProviderFailed       = errors.New("provider driver operation failed")
	ErrProviderNotSupported = errors.New("provider cloud belum didukung atau driver belum terpasang, gunakan Mock Cloud Provider (slug: mock)")
	ErrServerNotRunning     = errors.New("server is not in running state")
	ErrServerNotStopped     = errors.New("server is not in stopped state")
)
