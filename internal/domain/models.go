package domain

// MergeRequest represents the merge request data used across the application.
type MergeRequest struct {
	ID              int
	Title           string
	State           string
	AuthorName      string
	SourceBranch    string
	MergeRequestURL string
}

// JiraIssue represents Jira task data used across the application.
type JiraIssue struct {
	Key       string
	Summary   string
	Host      string
	IssueType string
}

// Message is the combined output of processing a merge request URL.
type Message struct {
	ServiceName   string
	MergeRequest  MergeRequest
	JiraTask      JiraIssue
	HasMigrations bool
}
