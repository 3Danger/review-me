package shower

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"review-info/internal/pkg/gitlab"
	"review-info/internal/pkg/jira"
	"review-info/internal/pkg/shower/models"
)

type Service struct {
	gitlab *gitlab.Service
	jira   *jira.Service
}

func New(
	gitlab *gitlab.Service,
	jira *jira.Service,
) *Service {
	return &Service{
		gitlab: gitlab,
		jira:   jira,
	}
}

func split(urlStr string) (baeURL, projectPath string, id int, err error) {
	if _, err = url.Parse(urlStr); err != nil {
		return "", "", 0, fmt.Errorf("parsing url: %w", err)
	}

	reg, err := regexp.Compile(`^(https?://)(git.vseinstrumenti.net)(.*)(/-/merge_requests/)(\d{1,10})$`)
	if err != nil {
		return "", "", 0, err
	}

	res := reg.FindAllStringSubmatch(urlStr, -1)
	if len(res) != 1 || len(res[0]) != 6 {
		return "", "", 0, fmt.Errorf("invalid url: %s", urlStr)
	}

	idUint, err := strconv.ParseUint(res[0][5], 10, 64)
	if err != nil {
		return "", "", 0, fmt.Errorf("parsing id: %w", err)
	}

	return res[0][1] + res[0][2], res[0][3], int(idUint), nil
}

func (s *Service) Process(mergeRequestURL string) (*models.Message, error) {
	if mergeRequestURL == "" {
		return nil, errors.New("empty gitlab url")
	}

	_, path, mrID, err := split(mergeRequestURL)
	if err != nil {
		return nil, fmt.Errorf("splitting gitlab url: %w", err)
	}

	mr, err := s.gitlab.MergeRequest(path, mrID)
	if err != nil {
		return nil, fmt.Errorf("getting merge request: %w", err)
	}

	reg, err := regexp.Compile(`FD-\d{1,10}`)
	if err != nil {
		return nil, fmt.Errorf("compiling regexp: %w", err)
	}

	issue := reg.FindAllString(mr.SourceBranch, -1)
	if len(issue) != 1 {
		return nil, fmt.Errorf("invalid merge request branch: %s", issue)
	}

	task, err := s.jira.Get(issue[0])
	if err != nil {
		return nil, fmt.Errorf("getting task: %w", err)
	}

	return &models.Message{
		ServiceName: lastPart('/', path),
		MergeRequest: models.MergeRequest{
			ID:              mrID,
			MergeRequestURL: mergeRequestURL,
		},
		JiraTask: models.JiraTask{
			ID:      task.Key,
			Host:    "https://jsw.vseinstrumenti.ru/browse",
			Summary: task.Fields.Summary,
		},
	}, nil
}

func lastPart(delim rune, path string) string {
	path, _ = strings.CutSuffix(path, string(delim))
	splitted := strings.Split(path, "/")

	return splitted[len(splitted)-1]
}
