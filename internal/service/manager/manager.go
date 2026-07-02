package manager

import (
	"context"
	"fmt"
	"time"

	"review-info/internal/config"
	"review-info/internal/domain"
)

const (
	timeRoundStep = 15 * time.Minute
	deployWindow  = 30 * time.Minute
)

type Service struct {
	svc domain.MRProcessor
	cnf config.Message
}

// Ensure Service satisfies domain.ActionRunner at compile time.
var _ domain.ActionRunner = (*Service)(nil)

func New(cnf config.Message, svc domain.MRProcessor) *Service {
	return &Service{
		cnf: cnf,
		svc: svc,
	}
}

func (s *Service) Execute(ctx context.Context, action domain.ActionType, url string, opts domain.ActionOptions) (string, error) {
	switch action {
	case domain.ActionReview:
		return s.reviewMe(ctx, url)
	case domain.ActionDeploy:
		return s.deployPlanning(ctx, url, opts)
	default:
		return "", fmt.Errorf("unknown action: %s", action)
	}
}

func (s *Service) reviewMe(ctx context.Context, rawGitlabURL string) (string, error) {
	info, err := s.svc.Process(ctx, rawGitlabURL)
	if err != nil {
		return "", err
	}

	msg := s.cnf.Team +
		"\n" + s.cnf.Review +
		"\nСервис: " + info.ServiceName +
		"\n" + formatMessageShort(info)

	return msg, nil
}

func (s *Service) deployPlanning(ctx context.Context, rawGitlabURL string, opts domain.ActionOptions) (string, error) {
	info, err := s.svc.Process(ctx, rawGitlabURL)
	if err != nil {
		return "", err
	}

	loc := config.LoadLocation(opts.Timezone)
	now := time.Now().In(loc).Add(timeRoundStep).Truncate(timeRoundStep).Add(opts.After)

	msg := s.cnf.Team +
		"\n" + s.cnf.Deploy + ": с " + now.Format("15:04") +
		" по " + now.Add(deployWindow).Format("15:04") +
		"\nСервис: " + info.ServiceName +
		"\n" + formatMessageShort(info) +
		"\n" + migrationLine(info.HasMigrations, opts.MigrationsApplied)

	return msg, nil
}

func migrationLine(hasMigrations bool, migrationsApplied bool) string {
	if hasMigrations {
		if migrationsApplied {
			return "Миграции есть и применены в проде: :checkbox-selected:"
		}
		return "Миграции есть: :checkbox-empty:"
	}
	return "Миграций нет: :checkbox-selected:"
}

func formatMessageShort(msg *domain.Message) string {
	jiraLink := "[[" + msg.JiraTask.Key + "](" + msg.JiraTask.Host + "/" + msg.JiraTask.Key + ")]"
	mrLink := "[[MR-" + fmt.Sprint(msg.MergeRequest.ID) + "](" + msg.MergeRequest.MergeRequestURL + ")]"
	issueInfo := "[" + msg.JiraTask.IssueType + "] " + msg.JiraTask.Summary
	return jiraLink + " " + mrLink + " - " + issueInfo
}
