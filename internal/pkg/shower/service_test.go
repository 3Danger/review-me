package shower

import (
	"context"
	"testing"

	"review-info/internal/domain"
)

type stubGitLab struct {
	mr      *domain.MergeRequest
	changes *domain.MergeRequestChanges
}

func (s *stubGitLab) MergeRequest(_ context.Context, _ string, _ int) (*domain.MergeRequest, error) {
	return s.mr, nil
}

func (s *stubGitLab) MergeRequestChanges(_ context.Context, _ string, _ int) (*domain.MergeRequestChanges, error) {
	if s.changes == nil {
		return &domain.MergeRequestChanges{}, nil
	}
	return s.changes, nil
}

type stubJira struct {
	gotKey string
}

func (s *stubJira) Get(_ context.Context, issueKey string) (*domain.JiraIssue, error) {
	s.gotKey = issueKey
	return &domain.JiraIssue{Key: issueKey}, nil
}

const (
	testHost   = "git.example.com"
	testPrefix = "FD-"
	testMRURL  = "https://git.example.com/group/proj/-/merge_requests/5"
)

func TestResolveJiraKey(t *testing.T) {
	tests := []struct {
		name    string
		branch  string
		title   string
		wantKey string
		wantErr bool
	}{
		{
			name:    "key from branch",
			branch:  "feature/FD-1234-login",
			title:   "Login feature",
			wantKey: "FD-1234",
		},
		{
			name:    "key from title when branch has none",
			branch:  "hotfix/login-fix",
			title:   "FD-5678: fix login",
			wantKey: "FD-5678",
		},
		{
			name:    "branch wins over title",
			branch:  "feature/FD-1111-x",
			title:   "FD-2222 other",
			wantKey: "FD-1111",
		},
		{
			name:    "no key anywhere",
			branch:  "hotfix/login-fix",
			title:   "fix login",
			wantErr: true,
		},
		{
			name:    "ambiguous branch with two keys",
			branch:  "FD-1/FD-2-merge",
			title:   "FD-9999",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jira := &stubJira{}
			gl := &stubGitLab{mr: &domain.MergeRequest{SourceBranch: tt.branch, Title: tt.title}}
			svc := New(gl, jira, testHost, testPrefix)

			msg, err := svc.Process(context.Background(), testMRURL)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (key=%q)", jira.gotKey)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if jira.gotKey != tt.wantKey {
				t.Errorf("jira key = %q, want %q", jira.gotKey, tt.wantKey)
			}
			if msg.JiraTask.Key != tt.wantKey {
				t.Errorf("message jira key = %q, want %q", msg.JiraTask.Key, tt.wantKey)
			}
		})
	}
}
