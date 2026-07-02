package gui

import (
	"errors"
	"strings"

	"review-info/internal/domain"
)

// FormatErrorMessage formats error messages in Russian with appropriate context
func FormatErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	switch {
	case errors.Is(err, domain.ErrUnauthorized):
		return "Ошибка авторизации: проверьте токены доступа в .env файле."
	case errors.Is(err, domain.ErrForbidden):
		return "Ошибка доступа: недостаточно прав для выполнения операции."
	case errors.Is(err, domain.ErrNotFound):
		return "Ошибка: ресурс не найден. Проверьте правильность MR URL."
	case errors.Is(err, domain.ErrBadRequest):
		return "Ошибка: неверный запрос. Проверьте правильность введенных данных."
	case errors.Is(err, domain.ErrServerError):
		return "Ошибка сервера: внутренняя ошибка сервера. Попробуйте позже."
	case errors.Is(err, domain.ErrNetwork):
		return "Ошибка сети: не удалось подключиться к серверу. Проверьте подключение к интернету."
	default:
		msg := err.Error()
		if strings.Contains(msg, "timeout") {
			return "Ошибка сети: превышено время ожидания ответа от сервера. Попробуйте еще раз."
		}
		return "Ошибка: " + msg
	}
}
