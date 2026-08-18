package logger

import (
	"log/slog"
	"os"
	"strings"
)

var defaultLogger *slog.Logger

// Init mengonfigurasi dan menginisialisasi structured logger global.
// Parameter level menentukan ambang batas pencatatan log ("debug", "info", "warn", "error").
// Parameter isDev menentukan format handler: true untuk text handler (stdout), false untuk JSON handler (stdout).
// Mengembalikan pointer *slog.Logger yang telah disetel sebagai default logger.
func Init(level string, isDev bool) *slog.Logger {
	var logLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn", "warning":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	var handler slog.Handler
	if isDev {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)
	return defaultLogger
}

// Get mengambil instance singleton logger yang sedang aktif.
// Mengembalikan pointer *slog.Logger aktif atau slog.Default() jika belum diinisialisasi.
func Get() *slog.Logger {
	if defaultLogger == nil {
		defaultLogger = slog.Default()
	}
	return defaultLogger
}

// Debug mencatat log pada level Debug.
// Parameter msg berisi pesan log utama.
// Parameter args berisi key-value pairs atribut kontekstual log.
func Debug(msg string, args ...any) {
	Get().Debug(msg, args...)
}

// Info mencatat log pada level Info.
// Parameter msg berisi pesan log utama.
// Parameter args berisi key-value pairs atribut kontekstual log.
func Info(msg string, args ...any) {
	Get().Info(msg, args...)
}

// Warn mencatat log pada level Warn.
// Parameter msg berisi pesan log utama.
// Parameter args berisi key-value pairs atribut kontekstual log.
func Warn(msg string, args ...any) {
	Get().Warn(msg, args...)
}

// Error mencatat log pada level Error.
// Parameter msg berisi pesan log utama.
// Parameter args berisi key-value pairs atribut kontekstual log.
func Error(msg string, args ...any) {
	Get().Error(msg, args...)
}
