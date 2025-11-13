package jira

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"review-info/internal/config"
	"review-info/internal/pkg/jira/models"
)

type Service struct {
	config config.Jira
	client client
}

type client interface {
	Do(req *http.Request) (*http.Response, error)
}

func New(client client, config config.Jira) *Service {
	return &Service{
		config: config,
		client: client,
	}
}

func (s *Service) Get(issueKey string) (*models.Jira, error) {
	url := fmt.Sprintf("%s/rest/api/2/issue/%s", s.config.BaseURL, issueKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании запроса: %v", err)
	}

	// Базовая аутентификация
	basicAuth := base64.StdEncoding.EncodeToString([]byte(s.config.User + ":" + s.config.Pass))
	req.Header.Add("Authorization", "Basic "+basicAuth)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)

		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var issue models.Jira
	if err = json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, fmt.Errorf("json decoding response: %w", err)
	}

	return &issue, nil
}
