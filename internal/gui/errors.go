package gui

import "strings"

// FormatErrorMessage formats error messages in Russian with appropriate context
func FormatErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	errMsg := err.Error()

	// Check for common error patterns and provide user-friendly messages
	switch {
	case strings.Contains(errMsg, "connection refused"):
		return "Ошибка сети: не удалось подключиться к серверу. Проверьте подключение к интернету."
	case strings.Contains(errMsg, "timeout"):
		return "Ошибка сети: превышено время ожидания ответа от сервера. Попробуйте еще раз."
	case strings.Contains(errMsg, "no such host"):
		return "Ошибка сети: не удалось найти сервер. Проверьте URL."
	case strings.Contains(errMsg, "401") || strings.Contains(errMsg, "unauthorized"):
		return "Ошибка авторизации: проверьте токены доступа в .env файле."
	case strings.Contains(errMsg, "403") || strings.Contains(errMsg, "forbidden"):
		return "Ошибка доступа: недостаточно прав для выполнения операции."
	case strings.Contains(errMsg, "404") || strings.Contains(errMsg, "not found"):
		return "Ошибка: ресурс не найден. Проверьте правильность MR URL."
	case strings.Contains(errMsg, "500") || strings.Contains(errMsg, "internal server error"):
		return "Ошибка сервера: внутренняя ошибка сервера. Попробуйте позже."
	case strings.Contains(errMsg, "bad request"):
		return "Ошибка: неверный запрос. Проверьте правильность введенных данных."
	default:
		// Return the original error message with "Ошибка:" prefix
		return "Ошибка: " + errMsg
	}
}
