# AGENTS.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

## Project overview
- This is a Go application (`module review-info`) that generates formatted team messages from a GitLab Merge Request by combining GitLab MR data and Jira issue data.
- It supports two entry modes from `main.go`:
  - **GUI mode** (default when no CLI flags are present) using Gio (`gioui.org`)
  - **CLI mode** (when `-cli` or CLI flags are passed)

## Common commands
- Build current platform binary:
  - `go build -o bin/review-info .`
- Cross-platform build artifacts (Makefile):
  - `make build-all`
  - Individual targets: `make build-windows`, `make build-macos-amd64`, `make build-macos-arm64`, `make build-linux`
- Clean build output:
  - `make clean`
- Run tests:
  - `go test ./...`
- Run a single package:
  - `go test ./internal/gui`
- Run a single test by name:
  - `go test ./internal/gui -run TestNew`
- Verbose test run:
  - `go test -v ./...`
- Basic static checks (no dedicated lint config in repo):
  - `go vet ./...`
  - `gofmt -w .`

## Runtime requirements and configuration
- The app expects a `config.yml` in the current working directory when starting (`config.Load("config.yml")` in `main.go`).
- `config.yml` must include:
  - `gitlab.baseURL`, `gitlab.token`
  - `jira.baseURL`, `jira.user`, `jira.password`
  - `message.team`, `message.review`, `message.deploy`
- GUI preferences are persisted per OS in JSON via `internal/preferences/preferences.go`:
  - macOS: `~/Library/Application Support/review-info/preferences.json`
  - Linux: `~/.config/review-info/preferences.json` (or `$XDG_CONFIG_HOME/review-info/preferences.json`)
  - Windows: `%APPDATA%\review-info\preferences.json`

## High-level architecture
### End-to-end flow
1. `main.go` chooses GUI/CLI mode.
2. Both modes create `manager.Service` (`internal/service/manager`) with message templates from config.
3. `manager.Service` delegates to `shower.Service` (`internal/pkg/shower`) for data aggregation.
4. `shower.Service.Process`:
   - parses MR URL and extracts project path + MR IID,
   - fetches MR from GitLab (`internal/pkg/gitlab/service.go`),
   - extracts Jira key (`FD-<digits>`) from `mr.SourceBranch`,
   - fetches Jira issue (`internal/pkg/jira/service.go`),
   - returns a unified `models.Message`.
5. `manager.Service` formats final text for review/deploy actions.

### Layer responsibilities
- `internal/pkg/gitlab`: thin GitLab API client (MR fetch by project path + IID).
- `internal/pkg/jira`: thin Jira API client (issue fetch by key via basic auth).
- `internal/pkg/shower`: orchestration + URL/branch parsing and domain message assembly.
- `internal/service/manager`: final human-readable message formatting and deploy time-window logic.
- `internal/gui`: Gio UI components/events, input validation, error text mapping, clipboard integration.
- `internal/config`: YAML config loading and Moscow timezone helper.
- `internal/preferences`: persisted GUI defaults (action/timezone/team).

## Important implementation notes
- URL handling in `shower.split` is currently strict: it matches `git.vseinstrumenti.net` merge request URLs only. If behavior needs to support other hosts, update regex logic there first.
- GUI-level MR URL validation (`internal/gui/validation.go`) is intentionally looser than backend parsing; backend (`shower.split`) is source of truth.
- `internal/pkg/gitlab/service_test.go` and parts of `internal/integration_test.go` use real network/config and may fail in isolated CI/local environments without valid credentials and reachable services.
- Deploy text timing differs by API:
  - CLI uses `DeployPlaning(..., 30*time.Minute)` (offset start).
  - GUI uses `DeployPlaningWithTimezone(..., 0, timezone)` (immediate quarter-hour rounding in selected timezone).
