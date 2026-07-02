package jira

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"strings"

	"review-info/internal/config"
	"review-info/internal/domain"
	"review-info/internal/pkg/httpclient"
	"review-info/internal/pkg/jira/models"
)

// Ensure Service satisfies domain.JiraClient at compile time.
var _ domain.JiraClient = (*Service)(nil)

type Service struct {
	client  *httpclient.Client
	user    string
	pass    string
	baseURL string
}

func New(client domain.HTTPClient, cfg config.Jira) *Service {
	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	return &Service{
		client:  httpclient.New(client, baseURL),
		user:    cfg.User,
		pass:    cfg.Pass,
		baseURL: baseURL,
	}
}

func (s *Service) Get(ctx context.Context, issueKey string) (*domain.JiraIssue, error) {
	path := fmt.Sprintf("/rest/api/2/issue/%s", issueKey)
	auth := base64.StdEncoding.EncodeToString([]byte(s.user + ":" + s.pass))
	headers := map[string]string{"Authorization": "Basic " + auth}

	var issue models.Jira
	if err := s.client.Get(ctx, path, headers, &issue); err != nil {
		slog.Error("fetching jira issue", "component", "jira", "operation", "get_issue", "issue_key", issueKey, "error", err)
		return nil, err
	}

	return &domain.JiraIssue{
		Key:       issue.Key,
		Summary:   issue.Fields.Summary,
		IssueType: issue.Fields.Issuetype.Name,
		Host:      s.baseURL + "/browse",
	}, nil
}
