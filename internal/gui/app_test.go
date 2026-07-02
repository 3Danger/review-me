package gui

import (
	"fmt"
	"testing"
	"time"

	"review-info/internal/domain"
	"review-info/internal/preferences"
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

	// Verify controller was created
	if app.ctrl == nil {
		t.Fatal("ctrl was not initialized")
	}

	// Verify preferences were set
	if app.ctrl.prefs != prefs {
		t.Error("preferences were not set correctly")
	}

	// Verify form state was initialized from preferences
	if app.ctrl.form.Team != prefs.Team {
		t.Errorf("team not initialized from preferences: got %q, want %q", app.ctrl.form.Team, prefs.Team)
	}

	if app.ctrl.form.Action != prefs.Action {
		t.Errorf("action not initialized from preferences: got %q, want %q", app.ctrl.form.Action, prefs.Action)
	}

	if app.ctrl.form.Timezone != prefs.Timezone {
		t.Errorf("timezone not initialized from preferences: got %q, want %q", app.ctrl.form.Timezone, prefs.Timezone)
	}

	// Verify initial state
	if app.ctrl.form.MRURL != "" {
		t.Errorf("mrURL should be empty initially, got %q", app.ctrl.form.MRURL)
	}

	if app.ctrl.execution.Result() != "" {
		t.Errorf("result should be empty initially, got %q", app.ctrl.execution.Result())
	}

	if app.ctrl.execution.IsLoading() {
		t.Error("loading should be false initially")
	}

	if app.ctrl.execution.Error() != "" {
		t.Errorf("error should be empty initially, got %q", app.ctrl.execution.Error())
	}
}

func TestAppStructure(t *testing.T) {
	// Verify that App struct has all required fields
	prefs := &preferences.Preferences{
		Action:   "review",
		Timezone: "Europe/Moscow",
		Team:     "@test-team",
	}

	var svc domain.ActionRunner // nil is fine for structure test

	app := New(svc, prefs)

	// Test that all fields are accessible (compilation test)
	_ = app.window
	_ = app.theme
	_ = app.ctrl
	_ = app.ctrl.service
	_ = app.ctrl.form.MRURL
	_ = app.ctrl.form.Team
	_ = app.ctrl.form.Action
	_ = app.ctrl.form.Timezone
	_ = app.ctrl.execution.Result()
	_ = app.ctrl.execution.IsLoading()
	_ = app.ctrl.execution.Error()
	_ = app.ctrl.clipboardError
	_ = app.ctrl.mrURLError
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
		err      error
		expected string
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: "",
		},
		{
			name:     "unauthorized error",
			err:      fmt.Errorf("%w: status=401", domain.ErrUnauthorized),
			expected: "Ошибка авторизации: проверьте токены доступа в .env файле.",
		},
		{
			name:     "forbidden error",
			err:      fmt.Errorf("%w: status=403", domain.ErrForbidden),
			expected: "Ошибка доступа: недостаточно прав для выполнения операции.",
		},
		{
			name:     "not found error",
			err:      fmt.Errorf("%w: status=404", domain.ErrNotFound),
			expected: "Ошибка: ресурс не найден. Проверьте правильность MR URL.",
		},
		{
			name:     "bad request error",
			err:      fmt.Errorf("%w: status=400", domain.ErrBadRequest),
			expected: "Ошибка: неверный запрос. Проверьте правильность введенных данных.",
		},
		{
			name:     "server error",
			err:      fmt.Errorf("%w: status=500", domain.ErrServerError),
			expected: "Ошибка сервера: внутренняя ошибка сервера. Попробуйте позже.",
		},
		{
			name:     "network error",
			err:      fmt.Errorf("%w: connection refused", domain.ErrNetwork),
			expected: "Ошибка сети: не удалось подключиться к серверу. Проверьте подключение к интернету.",
		},
		{
			name:     "timeout error",
			err:      fmt.Errorf("request timeout"),
			expected: "Ошибка сети: превышено время ожидания ответа от сервера. Попробуйте еще раз.",
		},
		{
			name:     "generic error",
			err:      fmt.Errorf("something went wrong"),
			expected: "Ошибка: something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatErrorMessage(tt.err)
			if result != tt.expected {
				t.Errorf("FormatErrorMessage(%v) = %q, want %q", tt.err, result, tt.expected)
			}
		})
	}
}
