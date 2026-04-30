package gui

import (
	"testing"
	"time"

	"review-info/internal/preferences"
	"review-info/internal/service/manager"
)

func TestNew(t *testing.T) {
	// Create mock preferences
	prefs := &preferences.Preferences{
		Action:      "review",
		Timezone:    "Europe/Moscow",
		Team:        "@test-team",
		LastUpdated: time.Now(),
	}

	// Create app with nil service (we're just testing structure)
	app := New(nil, prefs)

	// Verify app was created
	if app == nil {
		t.Fatal("New() returned nil")
	}

	// Verify theme was initialized
	if app.theme == nil {
		t.Error("theme was not initialized")
	}

	// Verify preferences were set
	if app.prefs != prefs {
		t.Error("preferences were not set correctly")
	}

	// Verify state was initialized from preferences
	if app.team != prefs.Team {
		t.Errorf("team not initialized from preferences: got %q, want %q", app.team, prefs.Team)
	}

	if app.action != prefs.Action {
		t.Errorf("action not initialized from preferences: got %q, want %q", app.action, prefs.Action)
	}

	if app.timezone != prefs.Timezone {
		t.Errorf("timezone not initialized from preferences: got %q, want %q", app.timezone, prefs.Timezone)
	}

	// Verify initial state
	if app.mrURL != "" {
		t.Errorf("mrURL should be empty initially, got %q", app.mrURL)
	}

	if app.result != "" {
		t.Errorf("result should be empty initially, got %q", app.result)
	}

	if app.loading {
		t.Error("loading should be false initially")
	}

	if app.error != "" {
		t.Errorf("error should be empty initially, got %q", app.error)
	}
}

func TestAppStructure(t *testing.T) {
	// Verify that App struct has all required fields
	prefs := &preferences.Preferences{
		Action:   "review",
		Timezone: "Europe/Moscow",
		Team:     "@test-team",
	}

	var svc *manager.Service // nil is fine for structure test

	app := New(svc, prefs)

	// Test that all fields are accessible (compilation test)
	_ = app.window
	_ = app.theme
	_ = app.prefs
	_ = app.service
	_ = app.mrURL
	_ = app.team
	_ = app.action
	_ = app.timezone
	_ = app.result
	_ = app.loading
	_ = app.error
	_ = app.clipboardError
	_ = app.mrURLError
}

func TestIsValidMRURL(t *testing.T) {
	tests := []struct {
		name  string
		url   string
		valid bool
	}{
		{
			name:  "valid https merge_requests URL",
			url:   "https://gitlab.com/project/repo/-/merge_requests/123",
			valid: true,
		},
		{
			name:  "valid http merge_requests URL",
			url:   "http://gitlab.example.com/project/-/merge_requests/456",
			valid: true,
		},
		{
			name:  "valid merge_request singular",
			url:   "https://gitlab.com/project/-/merge_request/789",
			valid: true,
		},
		{
			name:  "empty URL",
			url:   "",
			valid: false,
		},
		{
			name:  "URL without protocol",
			url:   "gitlab.com/project/-/merge_requests/123",
			valid: false,
		},
		{
			name:  "URL without merge_requests",
			url:   "https://gitlab.com/project/repo",
			valid: false,
		},
		{
			name:  "URL with only whitespace",
			url:   "   ",
			valid: false,
		},
		{
			name:  "URL with merge_requests but no protocol",
			url:   "gitlab.com/merge_requests/123",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMRURL(tt.url)
			result := err == nil
			if result != tt.valid {
				t.Errorf("ValidateMRURL(%q) = %v, want %v", tt.url, result, tt.valid)
			}
		})
	}
}

func TestFormatErrorMessage(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected string
	}{
		{
			name:     "connection refused error",
			errMsg:   "connection refused",
			expected: "Ошибка сети: не удалось подключиться к серверу. Проверьте подключение к интернету.",
		},
		{
			name:     "timeout error",
			errMsg:   "request timeout",
			expected: "Ошибка сети: превышено время ожидания ответа от сервера. Попробуйте еще раз.",
		},
		{
			name:     "no such host error",
			errMsg:   "no such host",
			expected: "Ошибка сети: не удалось найти сервер. Проверьте URL.",
		},
		{
			name:     "401 unauthorized",
			errMsg:   "401 unauthorized",
			expected: "Ошибка авторизации: проверьте токены доступа в .env файле.",
		},
		{
			name:     "403 forbidden",
			errMsg:   "403 forbidden",
			expected: "Ошибка доступа: недостаточно прав для выполнения операции.",
		},
		{
			name:     "404 not found",
			errMsg:   "404 not found",
			expected: "Ошибка: ресурс не найден. Проверьте правильность MR URL.",
		},
		{
			name:     "500 internal server error",
			errMsg:   "500 internal server error",
			expected: "Ошибка сервера: внутренняя ошибка сервера. Попробуйте позже.",
		},
		{
			name:     "bad request",
			errMsg:   "bad request",
			expected: "Ошибка: неверный запрос. Проверьте правильность введенных данных.",
		},
		{
			name:     "generic error",
			errMsg:   "something went wrong",
			expected: "Ошибка: something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock error
			err := &mockError{msg: tt.errMsg}
			result := FormatErrorMessage(err)
			if result != tt.expected {
				t.Errorf("FormatErrorMessage(%q) = %q, want %q", tt.errMsg, result, tt.expected)
			}
		})
	}
}

// mockError is a simple error implementation for testing
type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}
