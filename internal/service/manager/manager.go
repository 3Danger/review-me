package manager

import (
	"errors"
	"time"

	"review-info/internal/config"
	"review-info/internal/pkg/shower"
)

type Service struct {
	svc *shower.Service
}

func New(svc *shower.Service) *Service {
	return &Service{
		svc: svc,
	}
}

var ErrTeamInvalid = errors.New("team is invalid")

func (s *Service) ReviewMe(rawGitlabURL, team string) (string, error) {
	if team == "" || []rune(team)[0] != '@' {
		return "", ErrTeamInvalid
	}

	info, err := s.svc.Process(rawGitlabURL)
	if err != nil {
		return "", err
	}

	msg := team +
		"\nПосмотрите пожалуйста МР, " + "Сервис: " + info.ServiceName +
		"\n" + info.Short()

	return msg, nil
}

func (s *Service) DeployPlaning(rawGitlabURL, team string, after time.Duration) (string, error) {
	if team == "" || []rune(team)[0] != '@' {
		return "", ErrTeamInvalid
	}

	info, err := s.svc.Process(rawGitlabURL)
	if err != nil {
		return "", err
	}

	now := time.Now().In(config.LocationMSK()).Truncate(time.Minute * 15).Add(after)

	msg := team + "\nПланирую выкатку " +
		"[с " + now.Format("15:04") +
		" по " + now.Add(time.Minute*30).Format("15:04") +
		"]: По сервису " + info.ServiceName +
		"\n" + info.Short()

	return msg, nil
}
