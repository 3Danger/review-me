package shower

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"review-info/internal/domain"
	jiramodels "review-info/internal/pkg/jira/models"
	"review-info/internal/pkg/shower/models"
)

// Pre-compiled regular expressions.
var (
	mergeRequestURLReg = regexp.MustCompile(`^(https?://)(git.vseinstrumenti.net)(.*)(/-/merge_requests/)(\d{1,10})$`)
	jiraKeyReg         = regexp.MustCompile(`FD-\d{1,10}`)
)

type Service struct {
	gitlab domain.GitLabClient
	jira   domain.JiraClient
}

func New(
	gitlab domain.GitLabClient,
	jira domain.JiraClient,
) *Service {
	return &Service{
		gitlab: gitlab,
		jira:   jira,
	}
}

func (s *Service) Process(mergeRequestURL string) (*models.Message, error) {
	if mergeRequestURL == "" {
		return nil, errors.New("empty gitlab url")
	}

	_, path, mrID, err := parseURL(mergeRequestURL)
	if err != nil {
		return nil, fmt.Errorf("splitting gitlab url: %w", err)
	}

	mr, err := s.gitlab.MergeRequest(path, mrID)
	if err != nil {
		return nil, fmt.Errorf("getting merge request: %w", err)
	}

	task, err := s.fetchJiraTaskFromBranch(mr.SourceBranch)
	if err != nil {
		return nil, fmt.Errorf("getting task: %w", err)
	}

	hasMigrations := s.detectMigrations(path, mrID)

	return &models.Message{
		ServiceName: lastPart('/', path),
		MergeRequest: models.MergeRequest{
			ID:              mrID,
			MergeRequestURL: mergeRequestURL,
		},
		JiraTask: models.JiraTask{
			ID:        task.Key,
			Host:      "https://jsw.vseinstrumenti.ru/browse",
			Summary:   task.Fields.Summary,
			IssueType: task.Fields.Issuetype.Name,
		},
		HasMigrations: hasMigrations,
	}, nil
}

// fetchJiraTaskFromBranch extracts a Jira issue key from the branch name and fetches the task.
func (s *Service) fetchJiraTaskFromBranch(branch string) (*jiramodels.Jira, error) {
	issue := jiraKeyReg.FindAllString(branch, -1)
	if len(issue) != 1 {
		return nil, fmt.Errorf("invalid merge request branch: %s", issue)
	}
	return s.jira.Get(issue[0])
}

// detectMigrations checks whether the merge request changes include migration files.
func (s *Service) detectMigrations(path string, mrID int) bool {
	changes, err := s.gitlab.MergeRequestChanges(path, mrID)
	if err != nil {
		return false
	}
	for _, change := range changes.Changes {
		if isMigrationFile(change.OldPath) || isMigrationFile(change.NewPath) {
			return true
		}
	}
	return false
}

func parseURL(urlStr string) (baseURL, projectPath string, id int, err error) {
	if _, err = url.Parse(urlStr); err != nil {
		return "", "", 0, fmt.Errorf("parsing url: %w", err)
	}

	res := mergeRequestURLReg.FindAllStringSubmatch(urlStr, -1)
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

func lastPart(delim rune, path string) string {
	path, _ = strings.CutSuffix(path, string(delim))
	splitted := strings.Split(path, "/")

	return splitted[len(splitted)-1]
}
