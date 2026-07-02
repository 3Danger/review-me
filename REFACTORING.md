# План рефакторинга — review-info

На основе [code review эссе](/.claude/plans/keen-launching-blum.md).

## Как пользоваться этим файлом

- Каждый раздел — **самостоятельный блок работы**, который можно делегировать отдельному агенту
- Разделы пронумерованы в порядке выполнения (нижние зависят от верхних)
- Внутри раздела подпункты выполняются последовательно
- После выполнения каждого раздела запускать `go build ./... && go vet ./... && go test ./...` (если тесты проходят до изменений)

---

## 1. Санитарный минимум (без зависимостей)

### 1.1 Удалить дубликат main.go

**Файл**: `main.go` (корень)
**Действие**: Удалить файл. Рабочий entrypoint — `cmd/review-me/main.go`.
**Проверка**: `go build ./...` не выдаёт ошибок.

### 1.2 HTTP-клиент с таймаутом

**Файлы**: `cmd/review-me/main.go` (строки 94, 136), `internal/integration_test.go` (строка 40)
**Действие**: Заменить `new(http.Client)` на `&http.Client{Timeout: 30 * time.Second}`.
**Проверка**: Сборка проходит.

### 1.3 Починить TestMain в gitlab/service_test.go

**Файл**: `internal/pkg/gitlab/service_test.go:24`
**Действие**: Заменить `t.Run()` на `os.Exit(m.Run())`.
**Проверка**: `go vet ./internal/pkg/gitlab/...` без ошибок.

### 1.4 Исправить порядок аргументов require.Equal

**Файл**: `internal/pkg/gitlab/service_test.go:37`
**Действие**: Поменять на `require.Equal(t, mrIID, mr.IID)`.
**Проверка**: `go vet` без ошибок.

---

## 2. Domain / DIP — переработать порты и модели

> **Зависит от**: завершён раздел 1

### 2.1 Определить доменные модели в domain/

**Файл**: `internal/domain/ports.go`
**Действие**:
1. Создать `internal/domain/models.go` с типами:
   ```go
   type MergeRequest struct {
       ID        int
       Title     string
       State     string
       CreatedAt time.Time
       UpdatedAt time.Time
       Author    string
       ProjectPath string
       MRURL     string
   }

   type JiraIssue struct {
       ID        string
       Key       string
       Summary   string
       IssueType string
       Host      string // base URL for link construction
   }

   type Message struct {
       ServiceName  string
       JiraTask     JiraIssue
       MergeRequest MergeRequest
   }
   ```
2. Убрать импорты `gitlabmodels`, `jiramodels`, `showermodels` из `domain/ports.go`
3. Интерфейсы в `ports.go` должны возвращать доменные типы, не pkg-типы
4. `ActionOptions.After` + `MigrationsApplied` оставить только в ActionOptions для deploy (или сделать отдельный тип `DeployOptions`)

### 2.2 Переделать реализации под доменные модели

**Файлы**: `internal/pkg/gitlab/service.go`, `internal/pkg/gitlab/models/models.go`, `internal/pkg/jira/service.go`, `internal/pkg/jira/models/jira.go`, `internal/pkg/shower/service.go`, `internal/pkg/shower/models/models.go`

**Действие**:
1. `gitlab.Service.MergeRequest()` — маппит `gitlabmodels.MergeRequest` → `domain.MergeRequest`
2. `jira.Service.Get()` — маппит `jiramodels.Jira` → `domain.JiraIssue`
3. `shower.Service.Process()` — возвращает `*domain.Message`
4. Удалить `shower/models/models.go` (заменён на domain модели)

### 2.3 Добавить compile-time проверки интерфейсов

**Файлы**: `internal/pkg/gitlab/service.go`, `internal/pkg/jira/service.go`, `internal/pkg/shower/service.go`, `internal/service/manager/manager.go`

**Действие**: Добавить в каждый implementation-пакет:
```go
var _ domain.GitLabClient = (*Service)(nil)
```

---

## 3. context.Context — добавить во все I/O функции

> **Зависит от**: завершён раздел 2 (чтобы не переписывать сигнатуры дважды)

### 3.1 Пробросить ctx через все service-слои

**Файлы**: 
- `internal/pkg/gitlab/service.go` — `MergeRequest(ctx, ...)`, `MergeRequestChanges(ctx, ...)`
- `internal/pkg/jira/service.go` — `Get(ctx, ...)`
- `internal/pkg/shower/service.go` — `Process(ctx, ...)`
- `internal/service/manager/manager.go` — `Execute(ctx, ...)`
- `internal/domain/ports.go` — все интерфейсы

**Действие**:
1. `context.Context` — первый параметр каждого метода интерфейсов в `ports.go`
2. Все реализации обновить
3. Внутри `gitlab/service.go` и `jira/service.go` использовать `http.NewRequestWithContext(ctx, ...)`
4. Пробросить ctx через `main.go` → `runGUI`/`runCLI` → `manager.Service.Execute()`

### 3.2 Пробросить ctx через config.Load() и preferences

**Файлы**: `internal/config/config.go`, `internal/preferences/preferences.go`
**Действие**: Добавить `ctx` первым параметром во все публичные функции.

---

## 4. DRY — общая HTTP-логика для GitLab и Jira

> **Зависит от**: завершён раздел 3 (сигнатуры с ctx)

### 4.1 Создать internal/pkg/httpclient

**Новый пакет**: `internal/pkg/httpclient/`

**Действие**:
```go
package httpclient

type Client struct {
    raw     domain.HTTPClient
    baseURL string
}

func New(client domain.HTTPClient, baseURL string) *Client

func (c *Client) Get(ctx context.Context, path string, headers map[string]string, target interface{}) error
```

- Вынести общий паттерн: URL → NewRequestWithContext → set headers → Do → check status → ReadAll (error body) → JSON decode
- Добавить `statusCode` к результату для кастомной обработки ошибок

### 4.2 Переделать gitlab и jira сервисы на httpclient

**Файлы**: `internal/pkg/gitlab/service.go`, `internal/pkg/jira/service.go`

**Действие**: Заменить дублированный HTTP код на вызовы `httpclient.Client.Get()`.

---

## 5. KISS — убрать избыточную сложность

> **Зависит от**: завершён раздел 3 (сигнатуры с ctx)

### 5.1 Удалить registry.go, заменить на switch

**Файлы**: `internal/service/manager/registry.go`, `internal/service/manager/manager.go`

**Действие**:
1. Удалить `registry.go`
2. Удалить `ActionRegistry` struct, `ActionHandler` type, `Register()`
3. В `manager.go` сделать `Execute()` с прямым switch по `action`

### 5.2 Заменить кастомный .env парсер

**Файл**: `internal/config/config.go:87-122`

**Действие**:
1. Удалить функцию `readEnvFile()`
2. Использовать `github.com/joho/godotenv` (добавить в `go.mod`)
3. `godotenv.Load(path)` + `os.Getenv()` для каждого поля Config

### 5.3 Переделать GetFilePath() на os.UserConfigDir()

**Файл**: `internal/preferences/preferences.go:79-117`

**Действие**:
```go
func GetFilePath() (string, error) {
    dir, err := os.UserConfigDir()
    if err != nil { return "", err }
    return filepath.Join(dir, "review-info", "preferences.json"), nil
}
```

### 5.4 Избавиться от init() с panic в config

**Файл**: `internal/config/config.go:12-20`

**Действие**:
1. Удалить `init()` и `timeLocationMSK`
2. Удалить getter `LocationMSK()`
3. Передавать `*time.Location` как параметр через config, или загружать лениво при первом использовании с возвратом ошибки

### 5.5 Типизированные константы для action

**Файл**: `internal/domain/ports.go`

**Действие**:
```go
type ActionType string

const (
    ActionReview ActionType = "review"
    ActionDeploy ActionType = "deploy"
)

type ActionRunner interface {
    Execute(ctx context.Context, action ActionType, url string, opts ActionOptions) (string, error)
}
```

Пробросить `ActionType` через `manager.go`, `controller.go`, `app.go`.

### 5.6 Тройной цикл os.Args → один проход

**Файл**: `cmd/review-me/main.go:38-64`

**Действие**: Слить три цикла по `os.Args[1:]` в один.

---

## 6. GUI — Controller и Layout

> **Зависит от**: завершён раздел 3 (чтобы service-слои имели ctx)

### 6.1 Исправить data race в Controller

**Файл**: `internal/gui/controller.go:191-210`

**Действие**:
1. Добавить `sync.Mutex` в Controller (поле `mu`)
2. Защитить `c.loading`, `c.error`, `c.result` мьютексом при записи и чтении
3. Или вынести эти поля в отдельную структуру `State` с методами-геттерами/сеттерами под мьютексом
4. Вызов `c.window.Invalidate()` — через `op.InvalidateCmd{}` или заинжектить callback вместо прямого доступа к Window

### 6.2 Guard от конкурентных handleGenerate

**Файл**: `internal/gui/controller.go`

**Действие**: Добавить `sync.WaitGroup` или канал/флаг для блокировки повторного запуска, пока выполняется предыдущий. Показывать сообщение "already running".

### 6.3 Хранить widget.Editor между кадрами

**Файл**: `internal/gui/components.go:141-144`

**Действие**: `widget.Editor` вынести в поле `OutputWithCopy`, инициализировать один раз в `NewOutputWithCopy()`, в `Layout()` только обновлять текст если изменился.

### 6.4 Invalidate в Layout

**Файл**: `internal/gui/components.go:118-129`

**Действие**: Вынести таймер (2s copy feedback, 3s error feedback) из Layout в Controller или в отдельный компонент с каналом. Layout должен быть pure function.

### 6.5 Декомпозировать Controller

**Файл**: `internal/gui/controller.go`

**Действие**: 
1. Вынести auto-trigger логику в отдельный тип/структуру `AutoTrigger`
2. Вынести состояние формы в `FormState`
3. Вынести состояние выполнения/результата в `ExecutionState`
4. Controller остаётся тонким оркестратором

### 6.6 Повторяющийся паттерн ошибок — DRY

**Файл**: `internal/gui/app.go` (строки 130-137, 139-146, 253-273)

**Действие**: Вынести метод `renderError(gtx, message string) layout.Dimensions` в helpers.go или app.go.

### 6.7 Нестандартный MIME type

**Файл**: `internal/gui/clipboard.go:57`

**Действие**: `"application/text"` → `"text/plain"`.

---

## 7. Конфигурация — вынести хардкод

> **Зависит от**: завершён раздел 2 (чтобы не переписывать сигнатуры дважды)

### 7.1 Вынести GitLab хост в конфиг

**Файлы**: `internal/pkg/shower/service.go:18`, `internal/config/config.go`

**Действие**: 
- Добавить `GitLabHost string` в `config.Config.Gitlab`
- Передавать в `shower.Service` через конструктор
- Регулярное выражение строить динамически на основе хоста

### 7.2 Вынести Jira project prefix в конфиг

**Файл**: `internal/pkg/shower/service.go:19`, `internal/config/config.go`

**Действие**:
- Добавить `JiraProjectPrefix string` в `config.Config.Jira`
- Передавать в `shower.Service` через конструктор
- Регулярное выражение строить динамически

### 7.3 Вынести Jira host в конфиг

**Файл**: `internal/pkg/shower/service.go:68`, `internal/config/config.go`

**Действие**: Использовать `config.Jira.BaseURL` вместо хардкода `"https://jsw.vseinstrumenti.ru/browse"`.

### 7.4 Убрать хардкод таймзон в controller

**Файл**: `internal/gui/controller.go:49-52`

**Действие**: Список таймзон вынести в конфиг или package-level переменную с возможностью переопределения.

---

## 8. Код-стайл и идиоматичность

> **Зависит от**: завершены разделы 2-7 (чтобы не было конфликтов)

### 8.1 Исправить `lastPart` — порядок параметров и remove-obfuscation

**Файл**: `internal/pkg/shower/service.go:120-125`

**Действие**: Поменять сигнатуру на `lastPart(path string) string` (разделитель всегда `/`).

### 8.2 Сохранять widget.Editor между кадрами (уже в 6.3)

### 8.3 Sentinel errors

**Действие**: Определить sentinel-ошибки в `domain/` для известных сценариев:
- `var ErrNotFound = errors.New("not found")`
- `var ErrUnauthorized = errors.New("unauthorized")`
- `var ErrNetwork = errors.New("network error")`

Маппить HTTP status codes на эти ошибки в `gitlab/service.go` и `jira/service.go`.

### 8.4 Унифицировать язык ошибок

**Действие**: Все error strings — на английском, в нижнем регистре, без точки в конце. Русские сообщения оставить только в user-facing слое (`gui/errors.go`, `gui/validation.go`).

---

## 9. GUI — Controller decomposition (детали)

> **Зависит от**: завершён раздел 6

### 9.1 Выделить FormState

**Файл**: создать `internal/gui/form.go` (или расширить controller.go)

```go
type FormState struct {
    mu                sync.Mutex
    mrURL             string
    team              string
    action            domain.ActionType
    timezone          string
    migrationsApplied bool
}
```

### 9.2 Выделить ExecutionState

**Файл**: создать `internal/gui/execution.go`

```go
type ExecutionState struct {
    mu      sync.Mutex
    loading bool
    result  string
    err     string
    running bool // guard concurrent execution
}
```

### 9.3 Выделить AutoTrigger

**Файл**: создать `internal/gui/autotrigger.go`

```go
type AutoTrigger struct {
    lastAction, lastTimezone string
    lastMigrations          bool
}

func (a *AutoTrigger) IsDirty(action, tz string, migrations bool) bool
func (a *AutoTrigger) Update(action, tz string, migrations bool)
```

---

## 10. Тесты — **[ГОТОВО]**

### 10.1 Моки для GitLab/Jira через httptest

**Файл**: `internal/pkg/gitlab/service_test.go`

**Действие**: 
- Переписать тесты, используя `httptest.NewServer()`
- Не зависеть от внешнего GitLab/Jira
- Тесты: успешный ответ, 401, 404, 500, пустой body, невалидный JSON

### 10.2 Тесты Controller

**Новый файл**: `internal/gui/controller_test.go`

**Действие**:
- Тесты на `handleGenerate` (с моком ActionRunner)
- Тесты на auto-trigger detection
- Тесты на валидацию

### 10.3 Починить мёртвые тесты

**Файлы**: `internal/gui/app_test.go`, `internal/integration_test.go`

**Действие**:
- `TestAppStructure` — удалить (проверки компиляции не нужны)
- `TestGUIInitialization` — либо переименовать и дополнить, либо удалить
- `TestPreferencesPersistence` — сделать реальный save/load в temp dir

> **Зависит от**: завершены разделы 1-9 (стабильный код для тестирования)

### 10.1 Моки для GitLab/Jira через httptest

**Файл**: `internal/pkg/gitlab/service_test.go`

**Действие**: 
- Переписать тесты, используя `httptest.NewServer()`
- Не зависеть от внешнего GitLab/Jira
- Тесты: успешный ответ, 401, 404, 500, пустой body, невалидный JSON

### 10.2 Тесты Controller

**Новый файл**: `internal/gui/controller_test.go`

**Действие**:
- Тесты на `handleGenerate` (с моком ActionRunner)
- Тесты на auto-trigger detection
- Тесты на валидацию

### 10.3 Починить мёртвые тесты

**Файлы**: `internal/gui/app_test.go`, `internal/integration_test.go`

**Действие**:
- `TestAppStructure` — удалить (проверки компиляции не нужны)
- `TestGUIInitialization` — либо переименовать и дополнить, либо удалить
- `TestPreferencesPersistence` — сделать реальный save/load в temp dir

---

## 11. Makefile и CI — **[ГОТОВО]**

### 11.1 Свернуть build-рецепты

**Файл**: `Makefile`

**Действие**: Заменить 4 идентичных рецепта на цикл или pattern rule.

### 11.2 Добавить цели

**Файл**: `Makefile`

**Действие**: Добавить `test`, `vet`, `lint`, `tidy`, `clean`.

### 11.3 Убрать unused VERSION

**Файл**: `Makefile:5`

**Действие**: Либо передавать через `-ldflags`, либо удалить.

### 11.1 Свернуть build-рецепты

**Файл**: `Makefile`

**Действие**: Заменить 4 идентичных рецепта на цикл или pattern rule.

### 11.2 Добавить цели

**Файл**: `Makefile`

**Действие**: Добавить `test`, `vet`, `lint`, `tidy`, `clean`.

### 11.3 Убрать unused VERSION

**Файл**: `Makefile:5`

**Действие**: Либо передавать через `-ldflags`, либо удалить.

---

## 12. Сохраняемость приложения / Выгрузка — **[ГОТОВО]**

- Обновить `go.mod` если нужно (добавить `github.com/joho/godotenv`, обновить версии)
- `go mod tidy`
- `go vet ./...`
- `go build ./...`
- `go test ./...`

---

## Граф зависимостей разделов

```
1 (санминимум)
  ↓
2 (domain/DIP) ←── 3 (context) ←── 4 (httpclient)
  ↓                ↓                ↓
5 (KISS)          └── 6 (GUI fix) ──┘
  ↓                     ↓
7 (config)            8 (codestyle)
  ↓                     ↓
  └──── 10 (tests) ←────┘
           ↓
         11 (makefile/ci)
```

**Правило**: раздел можно начинать, когда все его зависимости завершены.
Два независимых раздела можно делать параллельно разными агентами.
