package shower

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"review-info/internal/domain"
)

type Service struct {
	gitlab     domain.GitLabClient
	jira       domain.JiraClient
	gitlabHost string
	jiraPrefix string
}

func New(
	gitlab domain.GitLabClient,
	jira domain.JiraClient,
	gitlabHost string,
	jiraPrefix string,
) *Service {
	return &Service{
		gitlab:     gitlab,
		jira:       jira,
		gitlabHost: gitlabHost,
		jiraPrefix: jiraPrefix,
	}
}

// Ensure Service satisfies domain.MRProcessor at compile time.
var _ domain.MRProcessor = (*Service)(nil)

// buildPatterns returns the merge request URL regex and Jira key regex
// compiled from the configured host and prefix.
func (s *Service) buildPatterns() (*regexp.Regexp, *regexp.Regexp) {
	mrRegex := regexp.MustCompile(fmt.Sprintf(`^(https?://)(%s)(.*)(/-/merge_requests/)(\d{1,10})$`, regexp.QuoteMeta(s.gitlabHost)))
	jiraRegex := regexp.MustCompile(fmt.Sprintf(`%s\d{1,10}`, regexp.QuoteMeta(s.jiraPrefix)))
	return mrRegex, jiraRegex
}

func (s *Service) Process(ctx context.Context, mergeRequestURL string) (*domain.Message, error) {
	if mergeRequestURL == "" {
		return nil, errors.New("empty gitlab url")
	}

	mrRegex, _ := s.buildPatterns()
	_, path, mrID, err := s.parseURL(mergeRequestURL, mrRegex)
	if err != nil {
		slog.Error("parsing gitlab url", "component", "shower", "operation", "parse_url", "url", mergeRequestURL, "error", err)
		return nil, fmt.Errorf("splitting gitlab url: %w", err)
	}

	mr, err := s.gitlab.MergeRequest(ctx, path, mrID)
	if err != nil {
		return nil, fmt.Errorf("getting merge request: %w", err)
	}

	issueKey, err := s.resolveJiraKey(mr.SourceBranch, mr.Title)
	if err != nil {
		slog.Error("resolving jira key", "component", "shower", "operation", "resolve_jira_key", "branch", mr.SourceBranch, "title", mr.Title, "error", err)
		return nil, fmt.Errorf("getting task: %w", err)
	}

	task, err := s.jira.Get(ctx, issueKey)
	if err != nil {
		return nil, fmt.Errorf("getting task: %w", err)
	}

	hasMigrations := s.detectMigrations(ctx, path, mrID)

	return &domain.Message{
		ServiceName: lastPart(path),
		MergeRequest: domain.MergeRequest{
			ID:              mrID,
			MergeRequestURL: mergeRequestURL,
		},
		JiraTask: domain.JiraIssue{
			Key:       task.Key,
			Host:      task.Host,
			Summary:   task.Summary,
			IssueType: task.IssueType,
		},
		HasMigrations: hasMigrations,
	}, nil
}

// resolveJiraKey extracts a Jira issue key from the merge request branch name,
// falling back to the merge request title when the branch contains no key.
// The branch must contain exactly one key; multiple matches are ambiguous.
func (s *Service) resolveJiraKey(branch, title string) (string, error) {
	_, jiraRegex := s.buildPatterns()

	branchKeys := jiraRegex.FindAllString(branch, -1)
	switch {
	case len(branchKeys) == 1:
		return branchKeys[0], nil
	case len(branchKeys) > 1:
		return "", fmt.Errorf("ambiguous jira key in branch %q: %v", branch, branchKeys)
	}

	if titleKey := jiraRegex.FindString(title); titleKey != "" {
		return titleKey, nil
	}

	return "", fmt.Errorf("jira key not found in branch %q or title %q", branch, title)
}

// detectMigrations checks whether the merge request changes include migration files.
func (s *Service) detectMigrations(ctx context.Context, path string, mrID int) bool {
	changes, err := s.gitlab.MergeRequestChanges(ctx, path, mrID)
	if err != nil {
		slog.Warn("detecting migrations", "component", "shower", "operation", "detect_migrations", "project", path, "mr_iid", mrID, "error", err)
		return false
	}
	for _, change := range changes.Changes {
		if isMigrationFile(change.OldPath) || isMigrationFile(change.NewPath) {
			return true
		}
	}
	return false
}

func (s *Service) parseURL(urlStr string, mrRegex *regexp.Regexp) (baseURL, projectPath string, id int, err error) {
	if _, err = url.Parse(urlStr); err != nil {
		return "", "", 0, fmt.Errorf("parsing url: %w", err)
	}

	res := mrRegex.FindAllStringSubmatch(urlStr, -1)
	if len(res) != 1 || len(res[0]) != 6 {
		return "", "", 0, fmt.Errorf("invalid url: %s", urlStr)
	}

	idUint, err := strconv.ParseUint(res[0][5], 10, 64)
	if err != nil {
		return "", "", 0, fmt.Errorf("parsing id: %w", err)
	}

	return res[0][1] + res[0][2], res[0][3], int(idUint), nil
}

func isMigrationFile(path string) bool {
	return strings.Contains(path, "/migration/") && strings.HasSuffix(path, ".sql")
}

func lastPart(path string) string {
	path, _ = strings.CutSuffix(path, "/")
	splitted := strings.Split(path, "/")
	return splitted[len(splitted)-1]
}
