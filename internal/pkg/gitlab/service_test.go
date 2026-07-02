package gitlab_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"review-info/internal/config"
	"review-info/internal/domain"
	"review-info/internal/pkg/gitlab"
)

func TestMergeRequest_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if r.Header.Get("PRIVATE-TOKEN") != "test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		resp := map[string]interface{}{
			"id":           12345,
			"iid":          592,
			"title":        "Test MR",
			"state":        "opened",
			"created_at":   "2024-01-01T10:00:00.000Z",
			"updated_at":   "2024-01-01T11:00:00.000Z",
			"author":       map[string]interface{}{"name": "Test User"},
			"source_branch": "feature/FD-1234-test",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	svc := gitlab.New(srv.Client(), config.Gitlab{
		BaseURL: srv.URL,
		Token:   "test-token",
	})

	mr, err := svc.MergeRequest(context.Background(), "fd/account-balance", 592)
	require.NoError(t, err)
	require.Equal(t, 592, mr.ID)
	require.Equal(t, "Test MR", mr.Title)
	require.Equal(t, "opened", mr.State)
	require.Equal(t, "feature/FD-1234-test", mr.SourceBranch)
}

func TestMergeRequest_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, `{"message": "404 Not found"}`)
	}))
	defer srv.Close()

	svc := gitlab.New(srv.Client(), config.Gitlab{
		BaseURL: srv.URL,
		Token:   "test-token",
	})

	_, err := svc.MergeRequest(context.Background(), "fd/unknown", 999)
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrNotFound)
}

func TestMergeRequest_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, `{"message": "401 Unauthorized"}`)
	}))
	defer srv.Close()

	svc := gitlab.New(srv.Client(), config.Gitlab{
		BaseURL: srv.URL,
		Token:   "bad-token",
	})

	_, err := svc.MergeRequest(context.Background(), "fd/account-balance", 592)
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrUnauthorized)
}

func TestMergeRequest_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, `{"message": "Internal error"}`)
	}))
	defer srv.Close()

	svc := gitlab.New(srv.Client(), config.Gitlab{
		BaseURL: srv.URL,
		Token:   "test-token",
	})

	_, err := svc.MergeRequest(context.Background(), "fd/account-balance", 592)
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrServerError)
}

func TestMergeRequestChanges(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Contains(t, r.URL.Path, "merge_requests/592/changes")

		resp := map[string]interface{}{
			"changes": []map[string]interface{}{
				{"old_path": "src/main.go", "new_path": "src/main.go", "new_file": false, "renamed_file": false, "deleted_file": false},
				{"old_path": "src/migration/001.sql", "new_path": "src/migration/001.sql", "new_file": true, "renamed_file": false, "deleted_file": false},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	svc := gitlab.New(srv.Client(), config.Gitlab{
		BaseURL: srv.URL,
		Token:   "test-token",
	})

	changes, err := svc.MergeRequestChanges(context.Background(), "fd/account-balance", 592)
	require.NoError(t, err)
	require.Len(t, changes.Changes, 2)
	require.True(t, changes.Changes[1].NewFile)
}

func TestMergeRequest_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `this is not json`)
	}))
	defer srv.Close()

	svc := gitlab.New(srv.Client(), config.Gitlab{
		BaseURL: srv.URL,
		Token:   "test-token",
	})

	_, err := svc.MergeRequest(context.Background(), "fd/account-balance", 592)
	require.Error(t, err)
}
