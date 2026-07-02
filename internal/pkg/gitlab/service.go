package gitlab

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"review-info/internal/config"
	"review-info/internal/domain"
	"review-info/internal/pkg/gitlab/models"
	"review-info/internal/pkg/httpclient"
)

// Ensure Service satisfies domain.GitLabClient at compile time.
var _ domain.GitLabClient = (*Service)(nil)

type Service struct {
	client  *httpclient.Client
	token   string
	baseURL string
}

func New(client domain.HTTPClient, cfg config.Gitlab) *Service {
	return &Service{
		client:  httpclient.New(client, cfg.BaseURL),
		token:   cfg.Token,
		baseURL: cfg.BaseURL,
	}
}

func (s *Service) MergeRequest(ctx context.Context, projectPath string, mrIID int) (*domain.MergeRequest, error) {
	path := fmt.Sprintf("/api/v4/projects/%s/merge_requests/%d", encodePath(projectPath), mrIID)
	headers := map[string]string{"PRIVATE-TOKEN": s.token}

	var mr models.MergeRequest
	if err := s.client.Get(ctx, path, headers, &mr); err != nil {
		slog.Error("fetching merge request", "component", "gitlab", "operation", "merge_request", "project", projectPath, "mr_iid", mrIID, "error", err)
		return nil, err
	}

	return mapMergeRequest(&mr, s.baseURL+path), nil
}

func (s *Service) MergeRequestChanges(ctx context.Context, projectPath string, mrIID int) (*domain.MergeRequestChanges, error) {
	path := fmt.Sprintf("/api/v4/projects/%s/merge_requests/%d/changes", encodePath(projectPath), mrIID)
	headers := map[string]string{"PRIVATE-TOKEN": s.token}

	var changes models.MergeRequestChanges
	if err := s.client.Get(ctx, path, headers, &changes); err != nil {
		slog.Error("fetching merge request changes", "component", "gitlab", "operation", "merge_request_changes", "project", projectPath, "mr_iid", mrIID, "error", err)
		return nil, err
	}

	return mapMergeRequestChanges(&changes), nil
}

func encodePath(projectPath string) string {
	projectPath = strings.TrimPrefix(projectPath, "/")
	return strings.ReplaceAll(projectPath, "/", "%2F")
}

func mapMergeRequest(mr *models.MergeRequest, mrURL string) *domain.MergeRequest {
	return &domain.MergeRequest{
		ID:              mr.IID,
		Title:           mr.Title,
		State:           mr.State,
		AuthorName:      mr.Author.Name,
		SourceBranch:    mr.SourceBranch,
		MergeRequestURL: mrURL,
	}
}

func mapMergeRequestChanges(changes *models.MergeRequestChanges) *domain.MergeRequestChanges {
	dc := make([]domain.Change, len(changes.Changes))
	for i, c := range changes.Changes {
		dc[i] = domain.Change{
			OldPath:     c.OldPath,
			NewPath:     c.NewPath,
			NewFile:     c.NewFile,
			RenamedFile: c.RenamedFile,
			DeletedFile: c.DeletedFile,
		}
	}
	return &domain.MergeRequestChanges{Changes: dc}
}
