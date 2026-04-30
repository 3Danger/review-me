package manager

import (
	"time"

	"review-info/internal/config"
	"review-info/internal/pkg/shower"
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
		"\n" + s.cnf.Review + ", " + "Сервис: " + info.ServiceName +
		"\n" + info.Short()

	return msg, nil
}

func (s *Service) DeployPlaning(rawGitlabURL string, after time.Duration) (string, error) {
	info, err := s.svc.Process(rawGitlabURL)
	if err != nil {
		return "", err
	}

	now := time.Now().In(config.LocationMSK()).Add(time.Minute * 15).Truncate(time.Minute * 15).Add(after)

	msg := s.cnf.Team +
		"\n" + s.cnf.Deploy + ": с " + now.Format("15:04") +
		" по " + now.Add(time.Minute*30).Format("15:04") +
		"\nСервис: " + info.ServiceName +
		"\n" + info.Short() +
		"\n" + migrationLine(info.HasMigrations)

	return msg, nil
}

func (s *Service) DeployPlaningWithTimezone(rawGitlabURL string, after time.Duration, timezone string) (string, error) {
	info, err := s.svc.Process(rawGitlabURL)
	if err != nil {
		return "", err
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return "", err
	}

	now := time.Now().In(loc).Truncate(time.Minute * 15).Add(after)

	msg := s.cnf.Team +
		"\n" + s.cnf.Deploy + ": с " + now.Format("15:04") +
		" по " + now.Add(time.Minute*30).Format("15:04") +
		"\nСервис: " + info.ServiceName +
		"\n" + info.Short() +
		"\n" + migrationLine(info.HasMigrations)

	return msg, nil
}

func migrationLine(hasMigrations bool) string {
	if hasMigrations {
		return "Миграции есть: :checkbox-empty:"
	}
	return "Миграций нет: :checkbox-selected:"
}
