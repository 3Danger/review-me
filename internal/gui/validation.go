package gui

import (
	"fmt"
	"strings"
)

// ValidateMRURL проверяет корректность MR URL
// Возвращает error с русским сообщением при невалидном URL
func ValidateMRURL(url string) error {
	url = strings.TrimSpace(url)
	if url == "" {
		return fmt.Errorf("Ошибка: MR URL не может быть пустым")
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("Ошибка: неверный формат MR URL. Ожидается URL вида: https://gitlab.../merge_requests/...")
	}

	if !strings.Contains(url, "merge_requests") && !strings.Contains(url, "merge_request") {
		return fmt.Errorf("Ошибка: неверный формат MR URL. Ожидается URL вида: https://gitlab.../merge_requests/...")
	}

	return nil
}

// ValidateNonEmpty проверяет что значение не пустое
// Возвращает error с русским сообщением при пустом значении
func ValidateNonEmpty(value string, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("Ошибка: %s не может быть пустым", fieldName)
	}
	return nil
}
