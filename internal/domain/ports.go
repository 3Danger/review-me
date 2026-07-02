package domain

import (
	"context"
	"errors"
	"net/http"
	"time"
)

var (
	ErrNotFound     = errors.New("resource not found")
	ErrUnauthorized = errors.New("unauthorized: check credentials")
	ErrForbidden    = errors.New("forbidden: insufficient permissions")
	ErrNetwork      = errors.New("network error")
	ErrBadRequest   = errors.New("bad request")
	ErrServerError  = errors.New("server error")
	ErrUnknown      = errors.New("unknown error")
)

// ActionType is a typed constant for supported actions.
type ActionType string

const (
	ActionReview ActionType = "review"
	ActionDeploy ActionType = "deploy"
)

// HTTPClient is the minimal HTTP client interface consumed by all API services.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// GitLabClient provides read operations against the GitLab merge request API.
type GitLabClient interface {
	MergeRequest(ctx context.Context, projectPath string, mrIID int) (*MergeRequest, error)
	MergeRequestChanges(ctx context.Context, projectPath string, mrIID int) (*MergeRequestChanges, error)
}

// JiraClient provides read operations against the Jira issue API.
type JiraClient interface {
	Get(ctx context.Context, issueKey string) (*JiraIssue, error)
}

// MRProcessor processes a merge request URL and returns structured information.
type MRProcessor interface {
	Process(ctx context.Context, mergeRequestURL string) (*Message, error)
}

// ActionOptions carries the parameters that differ across action types.
type ActionOptions struct {
	After             time.Duration
	Timezone          string
	MigrationsApplied bool
}

// ActionRunner executes a named action against a merge request URL.
type ActionRunner interface {
	Execute(ctx context.Context, action ActionType, url string, opts ActionOptions) (string, error)
}

// MergeRequestChanges holds the diff/file metadata from GitLab.
type MergeRequestChanges struct {
	Changes []Change
}

// Change represents a single file change in a merge request.
type Change struct {
	OldPath     string
	NewPath     string
	NewFile     bool
	RenamedFile bool
	DeletedFile bool
}
