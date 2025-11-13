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
		//mrURL  = os.Args[1]
		//team   = os.Args[2]
	)

	flag.Parse()

	svc := manager.New(
		shower.New(
			gitlab.New(client, cnf.Gitlab),
			jira.New(client, cnf.Jira),
		),
	)

	info, err := svc.ReviewMe(*mrURL, *team)
	if err != nil {
		return fmt.Errorf("format: %w", err)
	}

	// Вывод
	fmt.Println(info)

	info, err = svc.DeployPlaning(*mrURL, *team, time.Minute*30)
	if err != nil {
		return fmt.Errorf("format: %w", err)
	}

	// Вывод
	fmt.Println(info)

	return nil
}
