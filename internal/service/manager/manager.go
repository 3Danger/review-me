package manager

import (
	"time"

	"review-info/internal/config"
	"review-info/internal/domain"
)

const (
	timeRoundStep = 15 * time.Minute
	deployWindow  = 30 * time.Minute
)

type Service struct {
	svc      domain.MRProcessor
	cnf      config.Message
	registry *ActionRegistry
}

func New(cnf config.Message, svc domain.MRProcessor) *Service {
	s := &Service{
		cnf:      cnf,
		svc:      svc,
		registry: NewActionRegistry(),
	}
	s.registry.Register("review", s.reviewHandler)
	s.registry.Register("deploy", s.deployHandler)
	return s
}

func (s *Service) Execute(action string, url string, opts domain.ActionOptions) (string, error) {
	return s.registry.Execute(action, url, opts)
}

func (s *Service) reviewHandler(url string, _ domain.ActionOptions) (string, error) {
	return s.reviewMe(url)
}

func (s *Service) deployHandler(url string, opts domain.ActionOptions) (string, error) {
	return s.deployPlaningWithTimezone(url, opts.After, opts.Timezone, opts.MigrationsApplied)
}

func (s *Service) reviewMe(rawGitlabURL string) (string, error) {
	info, err := s.svc.Process(rawGitlabURL)
	if err != nil {
		return "", err
	}

	msg := s.cnf.Team +
		"\n" + s.cnf.Review +
		"\nСервис: " + info.ServiceName +
		"\n" + info.Short()

	return msg, nil
}

func (s *Service) deployPlaningWithTimezone(rawGitlabURL string, after time.Duration, timezone string, migrationsApplied bool) (string, error) {
	info, err := s.svc.Process(rawGitlabURL)
	if err != nil {
		return "", err
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return "", err
	}

	now := time.Now().In(loc).Add(timeRoundStep).Truncate(timeRoundStep).Add(after)

	msg := s.cnf.Team +
		"\n" + s.cnf.Deploy + ": с " + now.Format("15:04") +
		" по " + now.Add(deployWindow).Format("15:04") +
		"\nСервис: " + info.ServiceName +
		"\n" + info.Short() +
		"\n" + migrationLine(info.HasMigrations, migrationsApplied)

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
