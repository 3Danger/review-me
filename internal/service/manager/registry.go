package manager

import (
	"fmt"

	"review-info/internal/domain"
)

// ActionHandler is a function that executes a named action.
type ActionHandler func(url string, opts domain.ActionOptions) (string, error)

// ActionRegistry maps action names to their handlers.
type ActionRegistry struct {
	handlers map[string]ActionHandler
}

// NewActionRegistry creates an empty action registry.
func NewActionRegistry() *ActionRegistry {
	return &ActionRegistry{
		handlers: make(map[string]ActionHandler),
	}
}

// Register adds a handler for the given action name.
func (r *ActionRegistry) Register(name string, handler ActionHandler) {
	r.handlers[name] = handler
}

// Execute dispatches to the registered handler.
func (r *ActionRegistry) Execute(action string, url string, opts domain.ActionOptions) (string, error) {
	handler, ok := r.handlers[action]
	if !ok {
		return "", fmt.Errorf("unknown action: %s", action)
	}
	return handler(url, opts)
}
