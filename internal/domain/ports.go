package domain

import (
	"net/http"
	"time"

	gitlabmodels "review-info/internal/pkg/gitlab/models"
	jiramodels "review-info/internal/pkg/jira/models"
	showermodels "review-info/internal/pkg/shower/models"
)

// HTTPClient is the minimal HTTP client interface consumed by all API services.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// GitLabClient provides read operations against the GitLab merge request API.
type GitLabClient interface {
	MergeRequest(projectPath string, mrIID int) (*gitlabmodels.MergeRequest, error)
	MergeRequestChanges(projectPath string, mrIID int) (*gitlabmodels.MergeRequestChanges, error)
}

// JiraClient provides read operations against the Jira issue API.
type JiraClient interface {
	Get(issueKey string) (*jiramodels.Jira, error)
}

// MRProcessor processes a merge request URL and returns structured information.
type MRProcessor interface {
	Process(mergeRequestURL string) (*showermodels.Message, error)
}

// ActionOptions carries the parameters that differ across action types.
type ActionOptions struct {
	After             time.Duration
	Timezone          string
	MigrationsApplied bool
}

// ActionRunner executes a named action against a merge request URL.
type ActionRunner interface {
	Execute(action string, url string, opts ActionOptions) (string, error)
}
