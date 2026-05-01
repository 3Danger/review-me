package gitlab

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"review-info/internal/config"
	"review-info/internal/domain"
	"review-info/internal/pkg/gitlab/models"
)

type Service struct {
	client  domain.HTTPClient
	baseURL string
	token   string
}

func New(client domain.HTTPClient, config config.Gitlab) *Service {
	return &Service{
		client:  client,
		baseURL: config.BaseURL,
		token:   config.Token,
	}
}

func (s *Service) MergeRequest(projectPath string, mrIID int) (*models.MergeRequest, error) {
	projectPath = strings.TrimPrefix(projectPath, "/")
	encodedProjectPath := strings.ReplaceAll(projectPath, "/", "%2F")
	url := fmt.Sprintf("%s/api/v4/projects/%s/merge_requests/%d", s.baseURL, encodedProjectPath, mrIID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании запроса: %v", err)
	}

	req.Header.Add("PRIVATE-TOKEN", s.token)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("неудачный ответ API. Код состояния: %d, Тело: %s", resp.StatusCode, string(body))
	}

	var mr models.MergeRequest
	err = json.NewDecoder(resp.Body).Decode(&mr)
	if err != nil {
		return nil, fmt.Errorf("ошибка при декодировании ответа JSON: %v", err)
	}

	return &mr, nil
}

func (s *Service) MergeRequestChanges(projectPath string, mrIID int) (*models.MergeRequestChanges, error) {
	projectPath = strings.TrimPrefix(projectPath, "/")
	encodedProjectPath := strings.ReplaceAll(projectPath, "/", "%2F")
	url := fmt.Sprintf("%s/api/v4/projects/%s/merge_requests/%d/changes", s.baseURL, encodedProjectPath, mrIID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании запроса: %v", err)
	}

	req.Header.Add("PRIVATE-TOKEN", s.token)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("неудачный ответ API изменений. Код состояния: %d, Тело: %s", resp.StatusCode, string(body))
	}

	var changes models.MergeRequestChanges
	err = json.NewDecoder(resp.Body).Decode(&changes)
	if err != nil {
		return nil, fmt.Errorf("ошибка при декодировании ответа JSON изменений: %v", err)
	}

	return &changes, nil
}
