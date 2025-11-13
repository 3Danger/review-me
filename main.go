package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"review-info/internal/config"
	"review-info/internal/pkg/gitlab"
	"review-info/internal/pkg/jira"
	"review-info/internal/pkg/shower"
	"review-info/internal/service/manager"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	if len(os.Args) < 2 {
		os.Args = append(os.Args, "https://git.vseinstrumenti.net/fd/account-balance/-/merge_requests/591")
		//return fmt.Errorf("не передан аргумент")
	}

	cnf, err := config.Load("config.yml")
	if err != nil {
		panic(fmt.Errorf("load config: %w", err))
	}

	if err := run(cnf); err != nil {
		panic(err)
	}
}

func run(cnf *config.Config) error {
	var (
		client = new(http.Client)
		mrURL  = flag.String("mr", "", "URL of the merge request")
		team   = flag.String("team", "", "Team to notify (e.g., @team-name)")
		action = flag.String("action", "review", "what to do (deploy or review)")
	)

	flag.Parse()

	if *team != "" {
		cnf.Message.Team = *team
	}

	svc := manager.New(
		cnf.Message,
		shower.New(
			gitlab.New(client, cnf.Gitlab),
			jira.New(client, cnf.Jira),
		),
	)

	var (
		message string
		err     error
	)

	switch *action {
	case "review":
		message, err = svc.ReviewMe(*mrURL)
	case "deploy":
		message, err = svc.DeployPlaning(*mrURL, time.Minute*30)
	}

	if err != nil {
		return fmt.Errorf("format: %w", err)
	}

	// Вывод
	fmt.Println(message)

	return nil
}
