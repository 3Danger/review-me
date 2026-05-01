package gitlab_test

import (
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"review-info/internal/config"
	"review-info/internal/pkg/gitlab"
)

var svc *gitlab.Service

func TestMain(t *testing.M) {
	client := new(http.Client)

	cnf := config.Gitlab{
		BaseURL: "https://git.vseinstrumenti.net",
		Token:   os.Getenv("GITLAB_TOKEN"),
	}

	svc = gitlab.New(client, cnf) // Здесь должен быть реальный HTTP клиент и конфигурация GitLab
	t.Run()
}

func TestMergeRequest(t *testing.T) {
	if svc == nil {
		t.Fatal("Service is not initialized")
	}

	projectPath := "fd/account-balance"
	mrIID := 592

	mr, err := svc.MergeRequest(projectPath, mrIID)
	require.NoError(t, err)
	require.Equal(t, mr.IID, mrIID)
	require.Equal(t, mr.ProjectID, 1547)
	require.Equal(t, mr.State, "merged")
}
