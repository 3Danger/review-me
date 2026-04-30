package manager

import (
	"time"

	"review-info/internal/config"
	"review-info/internal/pkg/shower"
)

const (
	timeRoundStep = 15 * time.Minute
	deployWindow  = 30 * time.Minute
)

type Service struct {
	svc *shower.Service
	cnf config.Message
}

func New(cnf config.Message, svc *shower.Service) *Service {
	return &Service{
		cnf: cnf,
		svc: svc,
	}
}

func (s *Service) ReviewMe(rawGitlabURL string) (string, error) {
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

func (s *Service) DeployPlaning(rawGitlabURL string, after time.Duration, migrationsApplied bool) (string, error) {
	info, err := s.svc.Process(rawGitlabURL)
	if err != nil {
		return "", err
	}

	now := time.Now().In(config.LocationMSK()).Add(timeRoundStep).Truncate(timeRoundStep).Add(after)

	msg := s.cnf.Team +
		"\n" + s.cnf.Deploy + ": с " + now.Format("15:04") +
		" по " + now.Add(deployWindow).Format("15:04") +
		"\nСервис: " + info.ServiceName +
		"\n" + info.Short() +
		"\n" + migrationLine(info.HasMigrations, migrationsApplied)

	return msg, nil
}

func (s *Service) DeployPlaningWithTimezone(rawGitlabURL string, after time.Duration, timezone string, migrationsApplied bool) (string, error) {
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
