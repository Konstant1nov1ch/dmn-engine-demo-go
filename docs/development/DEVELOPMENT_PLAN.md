# 🚀 План разработки High-Performance DMN Engine на Go

## Обзор проекта

**Цель**: Создать высокопроизводительный DMN Engine на Go для high-load систем с микросервисной архитектурой.

**Стратегия MVP**: Использовать сторонний FEEL Engine (через REST/gRPC), сфокусироваться на:
- DMN парсинг и валидация
- Оркестрация Decision Requirements Graph (DRG)
- Кэширование и производительность
- Clean API

---

## 📋 Фаза 0: Исследование и подготовка (2-3 дня)

### 0.1 Выбор стороннего FEEL Engine

**Варианты:**

| Engine | Язык | Интеграция | Производительность | Лицензия |
|--------|------|------------|-------------------|----------|
| **feel-scala** (Camunda) | Scala/JVM | REST wrapper или GraalVM | Средняя | Apache 2.0 |
| **DSNTK** | Rust | Native binary + REST | Высокая | MIT |
| **Kogito** (RedHat) | Java | REST API | Средняя | Apache 2.0 |
| **js-feel** | JavaScript | V8 embed или REST | Низкая | MIT |

**Рекомендация для MVP**: 
- **feel-scala** с обёрткой в Docker — проверенное решение от Camunda
- **Альтернатива**: DSNTK (Rust) — очень быстрый, но менее распространён

### 0.2 Задачи фазы 0

- [ ] Развернуть feel-scala в Docker локально
- [ ] Написать тестовый запрос к FEEL API
- [ ] Замерить базовую производительность (latency, throughput)
- [ ] Изучить DMN 1.3 спецификацию (особенно Decision Tables и Hit Policies)
- [ ] Создать репозиторий и базовую структуру проекта

### 0.3 Структура репозитория

```
dmn-engine-go/
├── cmd/
│   └── server/
│       └── main.go              # Точка входа
├── internal/
│   ├── api/
│   │   ├── handlers.go          # HTTP handlers
│   │   ├── middleware.go        # Auth, logging, etc.
│   │   └── routes.go            # Router setup
│   ├── dmn/
│   │   ├── parser.go            # XML парсер
│   │   ├── model.go             # DMN структуры данных
│   │   ├── validator.go         # Валидация DMN
│   │   └── compiler.go          # Подготовка к выполнению
│   ├── engine/
│   │   ├── engine.go            # Основной движок
│   │   ├── evaluator.go         # Логика выполнения
│   │   ├── hitpolicy.go         # Hit policies
│   │   └── drg.go               # DRG traversal
│   ├── feel/
│   │   ├── client.go            # Клиент к FEEL сервису
│   │   ├── cache.go             # Кэш FEEL выражений
│   │   └── pool.go              # Connection pool
│   ├── storage/
│   │   ├── repository.go        # Интерфейс хранилища
│   │   ├── postgres.go          # PostgreSQL реализация
│   │   └── memory.go            # In-memory для тестов
│   └── cache/
│       ├── cache.go             # Интерфейс кэша
│       └── redis.go             # Redis реализация
├── pkg/
│   └── dmnxml/                  # Публичные типы для DMN XML
├── deployments/
│   ├── docker/
│   │   ├── Dockerfile
│   │   └── docker-compose.yml
│   └── k8s/
│       └── *.yaml
├── scripts/
│   ├── setup-feel-engine.sh
│   └── generate-mocks.sh
├── testdata/
│   └── dmn/                     # Тестовые DMN файлы
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## 📋 Фаза 1: DMN Parser (3-4 дня)

### 1.1 Цель
Парсить DMN 1.3 XML файлы в Go структуры.

### 1.2 Задачи

#### 1.2.1 Определить Go структуры для DMN модели

```go
// internal/dmn/model.go

package dmn

// Definitions - корневой элемент DMN
type Definitions struct {
    ID          string       `xml:"id,attr"`
    Name        string       `xml:"name,attr"`
    Namespace   string       `xml:"namespace,attr"`
    Decisions   []Decision   `xml:"decision"`
    InputData   []InputData  `xml:"inputData"`
    BusinessKnowledgeModels []BKM `xml:"businessKnowledgeModel"`
}

// Decision - решение
type Decision struct {
    ID                      string                   `xml:"id,attr"`
    Name                    string                   `xml:"name,attr"`
    InformationRequirements []InformationRequirement `xml:"informationRequirement"`
    DecisionTable           *DecisionTable           `xml:"decisionTable"`
    LiteralExpression       *LiteralExpression       `xml:"literalExpression"`
    Variable                *Variable                `xml:"variable"`
}

// DecisionTable - таблица решений
type DecisionTable struct {
    ID        string   `xml:"id,attr"`
    HitPolicy string   `xml:"hitPolicy,attr"`
    Inputs    []Input  `xml:"input"`
    Outputs   []Output `xml:"output"`
    Rules     []Rule   `xml:"rule"`
}

// ... остальные структуры
```

- [ ] Создать все структуры для DMN 1.3
- [ ] Поддержать Decision Tables
- [ ] Поддержать Literal Expressions
- [ ] Поддержать Information Requirements (связи между решениями)

#### 1.2.2 Реализовать парсер

```go
// internal/dmn/parser.go

package dmn

import (
    "encoding/xml"
    "io"
)

type Parser struct{}

func NewParser() *Parser {
    return &Parser{}
}

func (p *Parser) Parse(r io.Reader) (*Definitions, error) {
    var defs Definitions
    decoder := xml.NewDecoder(r)
    if err := decoder.Decode(&defs); err != nil {
        return nil, fmt.Errorf("failed to parse DMN XML: %w", err)
    }
    return &defs, nil
}

func (p *Parser) ParseFile(path string) (*Definitions, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer f.Close()
    return p.Parse(f)
}
```

- [ ] Реализовать базовый парсер
- [ ] Добавить обработку ошибок с указанием строки/позиции
- [ ] Написать unit тесты с реальными DMN файлами

#### 1.2.3 Реализовать валидатор

```go
// internal/dmn/validator.go

package dmn

type ValidationError struct {
    Field   string
    Message string
}

type Validator struct{}

func (v *Validator) Validate(defs *Definitions) []ValidationError {
    var errors []ValidationError
    
    // Проверка уникальности ID
    // Проверка наличия decision tables
    // Проверка корректности hit policies
    // Проверка циклических зависимостей в DRG
    
    return errors
}
```

- [ ] Валидация ID (уникальность, формат)
- [ ] Валидация Decision Tables (inputs/outputs/rules)
- [ ] Валидация Hit Policies
- [ ] Проверка циклов в DRG (топологическая сортировка)

### 1.3 Тестовые файлы

Создать набор тестовых DMN файлов:
- [ ] `simple_table.dmn` — простая таблица с UNIQUE policy
- [ ] `multi_output.dmn` — таблица с несколькими выходами
- [ ] `drg_hierarchy.dmn` — DRG с под-решениями
- [ ] `all_hit_policies.dmn` — примеры всех hit policies
- [ ] `literal_expression.dmn` — решения с FEEL выражениями

---

## 📋 Фаза 2: FEEL Client (2-3 дня)

### 2.1 Цель
Создать клиент для взаимодействия с внешним FEEL Engine.

### 2.2 Архитектура интеграции

```
┌─────────────────┐       HTTP/gRPC        ┌─────────────────┐
│  DMN Engine     │ ───────────────────▶   │  FEEL Engine    │
│  (Go)           │                        │  (feel-scala)   │
│                 │ ◀───────────────────   │                 │
└─────────────────┘       JSON Response    └─────────────────┘
```

### 2.3 Задачи

#### 2.3.1 Docker setup для FEEL Engine

```yaml
# deployments/docker/docker-compose.yml

version: '3.8'

services:
  feel-engine:
    image: camunda/feel-scala-repl:latest  # или custom image
    ports:
      - "8090:8090"
    environment:
      - JAVA_OPTS=-Xmx512m
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8090/health"]
      interval: 10s
      timeout: 5s
      retries: 3

  # Альтернатива: DSNTK
  # dsntk:
  #   image: dsntk/dsntk:latest
  #   ports:
  #     - "8090:8090"
```

- [ ] Создать Docker Compose для FEEL Engine
- [ ] Написать health check
- [ ] Создать скрипт быстрого запуска

#### 2.3.2 FEEL Client интерфейс

```go
// internal/feel/client.go

package feel

import (
    "context"
)

// EvaluateRequest - запрос на выполнение FEEL выражения
type EvaluateRequest struct {
    Expression string                 `json:"expression"`
    Context    map[string]interface{} `json:"context"`
}

// EvaluateResponse - результат выполнения
type EvaluateResponse struct {
    Result interface{} `json:"result"`
    Error  string      `json:"error,omitempty"`
}

// UnaryTestRequest - запрос на проверку unary test
type UnaryTestRequest struct {
    Expression string      `json:"expression"` // ">= 18", "[1..100]"
    Value      interface{} `json:"value"`
}

// Client - интерфейс FEEL клиента
type Client interface {
    // Evaluate выполняет FEEL выражение
    Evaluate(ctx context.Context, req *EvaluateRequest) (*EvaluateResponse, error)
    
    // EvaluateUnaryTest проверяет unary test
    EvaluateUnaryTest(ctx context.Context, req *UnaryTestRequest) (bool, error)
    
    // Health проверяет доступность сервиса
    Health(ctx context.Context) error
    
    // Close закрывает соединения
    Close() error
}
```

- [ ] Определить интерфейс Client
- [ ] Реализовать HTTP клиент
- [ ] Добавить retry logic с exponential backoff
- [ ] Добавить timeout configuration

#### 2.3.3 Connection Pool

```go
// internal/feel/pool.go

package feel

import (
    "sync"
    "net/http"
)

type PooledClient struct {
    clients []*http.Client
    mu      sync.Mutex
    index   int
}

func NewPooledClient(size int, baseURL string) *PooledClient {
    clients := make([]*http.Client, size)
    for i := 0; i < size; i++ {
        clients[i] = &http.Client{
            Timeout: 5 * time.Second,
            Transport: &http.Transport{
                MaxIdleConns:        100,
                MaxIdleConnsPerHost: 100,
                IdleConnTimeout:     90 * time.Second,
            },
        }
    }
    return &PooledClient{clients: clients}
}

func (p *PooledClient) getClient() *http.Client {
    p.mu.Lock()
    defer p.mu.Unlock()
    client := p.clients[p.index]
    p.index = (p.index + 1) % len(p.clients)
    return client
}
```

- [ ] Реализовать connection pool
- [ ] Добавить circuit breaker (для отказоустойчивости)
- [ ] Добавить метрики (latency, errors)

#### 2.3.4 Кэширование FEEL выражений

```go
// internal/feel/cache.go

package feel

import (
    "crypto/sha256"
    "encoding/hex"
    "sync"
    "time"
)

type CachedResult struct {
    Result    interface{}
    ExpiresAt time.Time
}

type ExpressionCache struct {
    cache map[string]*CachedResult
    mu    sync.RWMutex
    ttl   time.Duration
}

func (c *ExpressionCache) Get(expression string, context map[string]interface{}) (interface{}, bool) {
    key := c.makeKey(expression, context)
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    if result, ok := c.cache[key]; ok && time.Now().Before(result.ExpiresAt) {
        return result.Result, true
    }
    return nil, false
}

func (c *ExpressionCache) makeKey(expression string, context map[string]interface{}) string {
    // SHA256(expression + sorted(context))
    h := sha256.New()
    h.Write([]byte(expression))
    // ... serialize context
    return hex.EncodeToString(h.Sum(nil))
}
```

- [ ] Реализовать in-memory кэш
- [ ] Добавить Redis кэш для distributed setup
- [ ] Настроить TTL и eviction policy

---

## 📋 Фаза 3: Decision Engine Core (4-5 дней)

### 3.1 Цель
Реализовать логику выполнения DMN решений.

### 3.2 Задачи

#### 3.2.1 Engine структура

```go
// internal/engine/engine.go

package engine

type Engine struct {
    feelClient    feel.Client
    definitionRepo storage.DefinitionRepository
    cache         cache.Cache
    hitPolicies   map[string]HitPolicy
}

func NewEngine(
    feelClient feel.Client,
    repo storage.DefinitionRepository,
    cache cache.Cache,
) *Engine {
    e := &Engine{
        feelClient:     feelClient,
        definitionRepo: repo,
        cache:          cache,
        hitPolicies:    make(map[string]HitPolicy),
    }
    
    // Регистрация hit policies
    e.hitPolicies["UNIQUE"] = &UniqueHitPolicy{}
    e.hitPolicies["FIRST"] = &FirstHitPolicy{}
    e.hitPolicies["ANY"] = &AnyHitPolicy{}
    e.hitPolicies["PRIORITY"] = &PriorityHitPolicy{}
    e.hitPolicies["COLLECT"] = &CollectHitPolicy{}
    e.hitPolicies["RULE ORDER"] = &RuleOrderHitPolicy{}
    e.hitPolicies["OUTPUT ORDER"] = &OutputOrderHitPolicy{}
    
    return e
}
```

- [ ] Создать структуру Engine
- [ ] Добавить dependency injection
- [ ] Настроить конфигурацию

#### 3.2.2 DRG Traversal (топологическая сортировка)

```go
// internal/engine/drg.go

package engine

// BuildExecutionOrder возвращает порядок выполнения решений
func (e *Engine) BuildExecutionOrder(defs *dmn.Definitions, targetDecisionID string) ([]string, error) {
    // Строим граф зависимостей
    graph := make(map[string][]string) // decision -> dependencies
    
    for _, d := range defs.Decisions {
        deps := make([]string, 0)
        for _, req := range d.InformationRequirements {
            if req.RequiredDecision != nil {
                deps = append(deps, req.RequiredDecision.Href[1:]) // убираем #
            }
        }
        graph[d.ID] = deps
    }
    
    // Топологическая сортировка (Kahn's algorithm)
    return topologicalSort(graph, targetDecisionID)
}

func topologicalSort(graph map[string][]string, target string) ([]string, error) {
    // ... реализация
}
```

- [ ] Построение графа зависимостей
- [ ] Топологическая сортировка
- [ ] Обнаружение циклов
- [ ] Поддержка частичного выполнения (только нужные для target)

#### 3.2.3 Decision Table Evaluator

```go
// internal/engine/evaluator.go

package engine

type EvaluateRequest struct {
    DecisionKey string                 `json:"decisionKey"`
    Variables   map[string]interface{} `json:"variables"`
    TenantID    string                 `json:"tenantId,omitempty"`
}

type EvaluateResult struct {
    DecisionKey    string                   `json:"decisionKey"`
    Outputs        []map[string]interface{} `json:"outputs"`
    MatchedRules   []string                 `json:"matchedRules"`
    EvaluatedAt    time.Time                `json:"evaluatedAt"`
    DurationNs     int64                    `json:"durationNs"`
}

func (e *Engine) Evaluate(ctx context.Context, req *EvaluateRequest) (*EvaluateResult, error) {
    start := time.Now()
    
    // 1. Получить definition
    def, err := e.getDefinition(ctx, req.DecisionKey, req.TenantID)
    if err != nil {
        return nil, err
    }
    
    // 2. Построить порядок выполнения
    order, err := e.BuildExecutionOrder(def, req.DecisionKey)
    if err != nil {
        return nil, err
    }
    
    // 3. Выполнить решения в порядке зависимостей
    variables := copyMap(req.Variables)
    
    for _, decisionID := range order {
        decision := findDecision(def, decisionID)
        result, err := e.evaluateDecision(ctx, decision, variables)
        if err != nil {
            return nil, err
        }
        // Добавляем результат в контекст для следующих решений
        variables[decision.Variable.Name] = result
    }
    
    // 4. Вернуть результат целевого решения
    return &EvaluateResult{
        DecisionKey: req.DecisionKey,
        Outputs:     variables[req.DecisionKey].([]map[string]interface{}),
        DurationNs:  time.Since(start).Nanoseconds(),
    }, nil
}
```

- [ ] Реализовать основной метод Evaluate
- [ ] Выполнение отдельного Decision
- [ ] Выполнение Decision Table
- [ ] Выполнение Literal Expression
- [ ] Обработка ошибок и partial failures

#### 3.2.4 Hit Policies

```go
// internal/engine/hitpolicy.go

package engine

type MatchedRule struct {
    RuleID  string
    Outputs map[string]interface{}
}

type HitPolicy interface {
    Apply(matched []MatchedRule) ([]map[string]interface{}, error)
    ShouldStopOnFirstMatch() bool
}

// UniqueHitPolicy - только одно правило должно совпасть
type UniqueHitPolicy struct{}

func (p *UniqueHitPolicy) Apply(matched []MatchedRule) ([]map[string]interface{}, error) {
    if len(matched) > 1 {
        return nil, fmt.Errorf("UNIQUE hit policy violated: %d rules matched", len(matched))
    }
    if len(matched) == 0 {
        return nil, nil
    }
    return []map[string]interface{}{matched[0].Outputs}, nil
}

func (p *UniqueHitPolicy) ShouldStopOnFirstMatch() bool { return false }

// CollectHitPolicy - собирает все совпавшие правила
type CollectHitPolicy struct {
    Aggregation string // SUM, MIN, MAX, COUNT, or empty for list
}

func (p *CollectHitPolicy) Apply(matched []MatchedRule) ([]map[string]interface{}, error) {
    if p.Aggregation == "" {
        // Возвращаем список всех outputs
        results := make([]map[string]interface{}, len(matched))
        for i, m := range matched {
            results[i] = m.Outputs
        }
        return results, nil
    }
    
    // Применяем агрегацию
    switch p.Aggregation {
    case "SUM":
        // ... сумма по числовым выходам
    case "COUNT":
        return []map[string]interface{}{{"count": len(matched)}}, nil
    // ...
    }
    return nil, nil
}
```

- [ ] UNIQUE - только одно правило
- [ ] FIRST - первое совпавшее
- [ ] ANY - любое из совпавших (но все должны давать одинаковый результат)
- [ ] PRIORITY - по приоритету (порядок outputs)
- [ ] COLLECT - все совпавшие + агрегация (SUM, MIN, MAX, COUNT)
- [ ] RULE ORDER - все по порядку правил
- [ ] OUTPUT ORDER - все по порядку outputs

#### 3.2.5 Rule Evaluation

```go
// internal/engine/rule.go

package engine

func (e *Engine) evaluateRule(
    ctx context.Context,
    rule *dmn.Rule,
    inputs []dmn.Input,
    variables map[string]interface{},
) (bool, map[string]interface{}, error) {
    
    // Проверяем все input entries
    for i, entry := range rule.InputEntries {
        // Получаем значение из контекста
        inputExpr := inputs[i].InputExpression.Text
        value, err := e.feelClient.Evaluate(ctx, &feel.EvaluateRequest{
            Expression: inputExpr,
            Context:    variables,
        })
        if err != nil {
            return false, nil, err
        }
        
        // Проверяем unary test
        matched, err := e.feelClient.EvaluateUnaryTest(ctx, &feel.UnaryTestRequest{
            Expression: entry.Text,
            Value:      value.Result,
        })
        if err != nil {
            return false, nil, err
        }
        
        if !matched {
            return false, nil, nil // правило не совпало
        }
    }
    
    // Все условия выполнены - вычисляем outputs
    outputs := make(map[string]interface{})
    for i, entry := range rule.OutputEntries {
        result, err := e.feelClient.Evaluate(ctx, &feel.EvaluateRequest{
            Expression: entry.Text,
            Context:    variables,
        })
        if err != nil {
            return false, nil, err
        }
        outputs[rule.Outputs[i].Name] = result.Result
    }
    
    return true, outputs, nil
}
```

- [ ] Реализовать evaluateRule
- [ ] Оптимизировать batch запросы к FEEL
- [ ] Добавить параллельное выполнение правил (где возможно)

---

## 📋 Фаза 4: Storage Layer (2-3 дня)

### 4.1 Цель
Реализовать персистентное хранение definitions.

### 4.2 Задачи

#### 4.2.1 Repository Interface

```go
// internal/storage/repository.go

package storage

type DefinitionRepository interface {
    // Deploy сохраняет новую версию definition
    Deploy(ctx context.Context, def *Definition) error
    
    // GetByKey возвращает последнюю версию
    GetByKey(ctx context.Context, key string, tenantID string) (*Definition, error)
    
    // GetByKeyAndVersion возвращает конкретную версию
    GetByKeyAndVersion(ctx context.Context, key string, version int, tenantID string) (*Definition, error)
    
    // List возвращает список definitions
    List(ctx context.Context, filter *ListFilter) ([]*Definition, error)
    
    // Delete удаляет definition
    Delete(ctx context.Context, key string, tenantID string) error
}

type Definition struct {
    ID          string
    Key         string
    Version     int
    Name        string
    Source      string           // Original XML
    ParsedModel *dmn.Definitions // Parsed model
    Checksum    string
    TenantID    string
    CreatedAt   time.Time
}
```

- [ ] Определить интерфейс Repository
- [ ] Определить модель Definition

#### 4.2.2 In-Memory Implementation (для тестов)

```go
// internal/storage/memory.go

package storage

type MemoryRepository struct {
    definitions map[string]*Definition
    mu          sync.RWMutex
}

func NewMemoryRepository() *MemoryRepository {
    return &MemoryRepository{
        definitions: make(map[string]*Definition),
    }
}

func (r *MemoryRepository) Deploy(ctx context.Context, def *Definition) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    key := fmt.Sprintf("%s:%s:%d", def.TenantID, def.Key, def.Version)
    r.definitions[key] = def
    return nil
}

// ... остальные методы
```

- [ ] Реализовать in-memory repository
- [ ] Добавить версионирование
- [ ] Добавить multi-tenancy

#### 4.2.3 PostgreSQL Implementation

```go
// internal/storage/postgres.go

package storage

import (
    "github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
    pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
    return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Deploy(ctx context.Context, def *Definition) error {
    // Получаем следующую версию
    var nextVersion int
    err := r.pool.QueryRow(ctx, `
        SELECT COALESCE(MAX(version), 0) + 1 
        FROM dmn_definitions 
        WHERE key = $1 AND (tenant_id = $2 OR ($2 IS NULL AND tenant_id IS NULL))
    `, def.Key, def.TenantID).Scan(&nextVersion)
    if err != nil {
        return err
    }
    
    def.Version = nextVersion
    
    // Сохраняем
    _, err = r.pool.Exec(ctx, `
        INSERT INTO dmn_definitions (id, key, version, name, source, parsed_model, checksum, tenant_id)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `, def.ID, def.Key, def.Version, def.Name, def.Source, def.ParsedModel, def.Checksum, def.TenantID)
    
    return err
}
```

- [ ] Реализовать PostgreSQL repository
- [ ] Создать миграции
- [ ] Добавить индексы для производительности
- [ ] Тестирование с реальной БД

#### 4.2.4 Database Migrations

```sql
-- migrations/001_initial.up.sql

CREATE TABLE dmn_definitions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key         VARCHAR(255) NOT NULL,
    version     INT NOT NULL DEFAULT 1,
    name        VARCHAR(255),
    source      TEXT NOT NULL,
    parsed_model JSONB,
    checksum    VARCHAR(64),
    tenant_id   VARCHAR(64),
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(key, version, tenant_id)
);

CREATE INDEX idx_dmn_def_key ON dmn_definitions(key);
CREATE INDEX idx_dmn_def_tenant ON dmn_definitions(tenant_id) WHERE tenant_id IS NOT NULL;
CREATE INDEX idx_dmn_def_key_version ON dmn_definitions(key, version DESC);
```

- [ ] Создать начальную миграцию
- [ ] Настроить golang-migrate или аналог

---

## 📋 Фаза 5: REST API (2-3 дня)

### 5.1 Цель
Создать REST API для работы с DMN Engine.

### 5.2 Задачи

#### 5.2.1 Выбор HTTP фреймворка

**Варианты:**
- **Fiber** - быстрый, Express-like API
- **Echo** - минималистичный, хорошая документация
- **Chi** - stdlib-совместимый, легковесный
- **Gin** - популярный, много middleware

**Рекомендация**: Fiber или Chi для high-performance

#### 5.2.2 API Design

```yaml
# OpenAPI 3.0 spec

paths:
  /api/v1/definitions:
    post:
      summary: Deploy DMN definition
      requestBody:
        content:
          application/xml:
            schema:
              type: string
          multipart/form-data:
            schema:
              type: object
              properties:
                file:
                  type: string
                  format: binary
      responses:
        201:
          description: Definition deployed
          
    get:
      summary: List definitions
      parameters:
        - name: key
          in: query
        - name: tenantId
          in: query
      responses:
        200:
          description: List of definitions
          
  /api/v1/definitions/{key}:
    get:
      summary: Get definition by key
    delete:
      summary: Delete definition
      
  /api/v1/definitions/{key}/xml:
    get:
      summary: Get original XML
      
  /api/v1/evaluate:
    post:
      summary: Evaluate decision
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                decisionKey:
                  type: string
                variables:
                  type: object
      responses:
        200:
          description: Evaluation result
          
  /api/v1/evaluate/batch:
    post:
      summary: Batch evaluate decisions
      
  /health:
    get:
      summary: Health check
      
  /ready:
    get:
      summary: Readiness check
```

- [ ] Определить API endpoints
- [ ] Создать OpenAPI спецификацию

#### 5.2.3 Handlers Implementation

```go
// internal/api/handlers.go

package api

import (
    "github.com/gofiber/fiber/v2"
)

type Handler struct {
    engine *engine.Engine
    repo   storage.DefinitionRepository
}

// POST /api/v1/definitions
func (h *Handler) DeployDefinition(c *fiber.Ctx) error {
    // Получаем XML из body или файла
    var xmlContent []byte
    
    file, err := c.FormFile("file")
    if err == nil {
        f, _ := file.Open()
        xmlContent, _ = io.ReadAll(f)
    } else {
        xmlContent = c.Body()
    }
    
    // Парсим и валидируем
    parser := dmn.NewParser()
    defs, err := parser.Parse(bytes.NewReader(xmlContent))
    if err != nil {
        return c.Status(400).JSON(ErrorResponse{Error: err.Error()})
    }
    
    validator := dmn.NewValidator()
    if errors := validator.Validate(defs); len(errors) > 0 {
        return c.Status(400).JSON(ErrorResponse{Errors: errors})
    }
    
    // Сохраняем
    def := &storage.Definition{
        Key:         defs.Decisions[0].ID,
        Name:        defs.Name,
        Source:      string(xmlContent),
        ParsedModel: defs,
    }
    
    if err := h.repo.Deploy(c.Context(), def); err != nil {
        return c.Status(500).JSON(ErrorResponse{Error: err.Error()})
    }
    
    return c.Status(201).JSON(def)
}

// POST /api/v1/evaluate
func (h *Handler) Evaluate(c *fiber.Ctx) error {
    var req engine.EvaluateRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(ErrorResponse{Error: err.Error()})
    }
    
    result, err := h.engine.Evaluate(c.Context(), &req)
    if err != nil {
        return c.Status(500).JSON(ErrorResponse{Error: err.Error()})
    }
    
    return c.JSON(result)
}
```

- [ ] Реализовать все handlers
- [ ] Добавить валидацию входных данных
- [ ] Добавить error handling

#### 5.2.4 Middleware

```go
// internal/api/middleware.go

package api

// RequestID добавляет уникальный ID к каждому запросу
func RequestID() fiber.Handler {
    return func(c *fiber.Ctx) error {
        id := c.Get("X-Request-ID")
        if id == "" {
            id = uuid.New().String()
        }
        c.Set("X-Request-ID", id)
        c.Locals("requestId", id)
        return c.Next()
    }
}

// Logger логирует запросы
func Logger(logger *slog.Logger) fiber.Handler {
    return func(c *fiber.Ctx) error {
        start := time.Now()
        err := c.Next()
        
        logger.Info("request",
            "method", c.Method(),
            "path", c.Path(),
            "status", c.Response().StatusCode(),
            "duration_ms", time.Since(start).Milliseconds(),
            "request_id", c.Locals("requestId"),
        )
        
        return err
    }
}

// TenantID извлекает tenant из header или JWT
func TenantID() fiber.Handler {
    return func(c *fiber.Ctx) error {
        tenantID := c.Get("X-Tenant-ID")
        c.Locals("tenantId", tenantID)
        return c.Next()
    }
}
```

- [ ] Request ID middleware
- [ ] Logging middleware
- [ ] Tenant ID extraction
- [ ] Rate limiting (для production)
- [ ] CORS

---

## 📋 Фаза 6: Testing & Quality (3-4 дня)

### 6.1 Цель
Обеспечить качество кода и покрытие тестами.

### 6.2 Задачи

#### 6.2.1 Unit Tests

```go
// internal/dmn/parser_test.go

func TestParser_SimpleDecisionTable(t *testing.T) {
    xml := `<?xml version="1.0" encoding="UTF-8"?>
    <definitions xmlns="...">
        <decision id="test" name="Test Decision">
            <decisionTable hitPolicy="UNIQUE">
                <input id="input1">
                    <inputExpression typeRef="number">
                        <text>age</text>
                    </inputExpression>
                </input>
                <output id="output1" name="result" typeRef="string"/>
                <rule id="rule1">
                    <inputEntry><text>>= 18</text></inputEntry>
                    <outputEntry><text>"adult"</text></outputEntry>
                </rule>
            </decisionTable>
        </decision>
    </definitions>`
    
    parser := NewParser()
    defs, err := parser.Parse(strings.NewReader(xml))
    
    require.NoError(t, err)
    require.Len(t, defs.Decisions, 1)
    require.Equal(t, "test", defs.Decisions[0].ID)
    require.Equal(t, "UNIQUE", defs.Decisions[0].DecisionTable.HitPolicy)
}
```

- [ ] Тесты для Parser
- [ ] Тесты для Validator
- [ ] Тесты для каждого Hit Policy
- [ ] Тесты для DRG traversal
- [ ] Тесты для FEEL Client (mock)

#### 6.2.2 Integration Tests

```go
// internal/engine/engine_integration_test.go

func TestEngine_EvaluateWithFEEL(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    // Требует запущенного FEEL engine
    feelClient := feel.NewHTTPClient("http://localhost:8090")
    repo := storage.NewMemoryRepository()
    engine := NewEngine(feelClient, repo, nil)
    
    // Deploy test DMN
    xml := loadTestFile(t, "testdata/dmn/credit_decision.dmn")
    // ... deploy
    
    // Evaluate
    result, err := engine.Evaluate(context.Background(), &EvaluateRequest{
        DecisionKey: "creditDecision",
        Variables: map[string]interface{}{
            "applicant": map[string]interface{}{
                "age":    25,
                "income": 50000,
            },
        },
    })
    
    require.NoError(t, err)
    require.NotEmpty(t, result.Outputs)
}
```

- [ ] Integration tests с реальным FEEL engine
- [ ] Integration tests с PostgreSQL
- [ ] E2E API tests

#### 6.2.3 Benchmarks

```go
// internal/engine/engine_bench_test.go

func BenchmarkEngine_EvaluateSimple(b *testing.B) {
    engine := setupBenchEngine(b)
    req := &EvaluateRequest{
        DecisionKey: "simpleDecision",
        Variables:   map[string]interface{}{"x": 42},
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := engine.Evaluate(context.Background(), req)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkEngine_EvaluateComplex(b *testing.B) {
    // DMN с 100 правилами
}

func BenchmarkEngine_EvaluateDRG(b *testing.B) {
    // DMN с 5 под-решениями
}
```

- [ ] Benchmark простого решения
- [ ] Benchmark сложной таблицы (100+ правил)
- [ ] Benchmark DRG
- [ ] Профилирование (pprof)

#### 6.2.4 Test Data

Создать набор тестовых DMN файлов:

```
testdata/dmn/
├── simple/
│   ├── unique_hitpolicy.dmn
│   ├── first_hitpolicy.dmn
│   ├── collect_hitpolicy.dmn
│   └── literal_expression.dmn
├── complex/
│   ├── large_table_100_rules.dmn
│   ├── multi_output.dmn
│   └── nested_context.dmn
├── drg/
│   ├── two_level_hierarchy.dmn
│   ├── diamond_dependency.dmn
│   └── parallel_decisions.dmn
└── invalid/
    ├── missing_output.dmn
    ├── cyclic_dependency.dmn
    └── invalid_hitpolicy.dmn
```

- [ ] Создать тестовые DMN файлы
- [ ] Файлы для всех hit policies
- [ ] Файлы с ошибками для негативных тестов

---

## 📋 Фаза 7: Observability (2-3 дня)

### 7.1 Цель
Добавить метрики, логирование, трейсинг.

### 7.2 Задачи

#### 7.2.1 Metrics (Prometheus)

```go
// internal/metrics/metrics.go

package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    EvaluationDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "dmn_evaluation_duration_seconds",
            Help:    "Duration of decision evaluation",
            Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
        },
        []string{"decision_key", "hit_policy"},
    )
    
    EvaluationTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "dmn_evaluation_total",
            Help: "Total number of evaluations",
        },
        []string{"decision_key", "status"},
    )
    
    FEELCallDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "feel_call_duration_seconds",
            Help:    "Duration of FEEL engine calls",
            Buckets: []float64{.0005, .001, .005, .01, .025, .05},
        },
        []string{"operation"},
    )
    
    CacheHits = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "dmn_cache_hits_total",
            Help: "Cache hits",
        },
        []string{"cache_type"},
    )
)
```

- [ ] Метрики выполнения решений
- [ ] Метрики FEEL вызовов
- [ ] Метрики кэша
- [ ] Prometheus endpoint

#### 7.2.2 Structured Logging

```go
// internal/logging/logger.go

package logging

import (
    "log/slog"
    "os"
)

func NewLogger(level string) *slog.Logger {
    var logLevel slog.Level
    switch level {
    case "debug":
        logLevel = slog.LevelDebug
    case "info":
        logLevel = slog.LevelInfo
    case "warn":
        logLevel = slog.LevelWarn
    case "error":
        logLevel = slog.LevelError
    default:
        logLevel = slog.LevelInfo
    }
    
    handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: logLevel,
    })
    
    return slog.New(handler)
}

// Логирование выполнения
func LogEvaluation(logger *slog.Logger, req *EvaluateRequest, result *EvaluateResult, err error) {
    if err != nil {
        logger.Error("evaluation failed",
            "decision_key", req.DecisionKey,
            "error", err.Error(),
        )
        return
    }
    
    logger.Info("evaluation completed",
        "decision_key", req.DecisionKey,
        "matched_rules", len(result.MatchedRules),
        "duration_ns", result.DurationNs,
    )
}
```

- [ ] Structured JSON logging
- [ ] Log levels configuration
- [ ] Request/response logging

#### 7.2.3 Distributed Tracing (OpenTelemetry)

```go
// internal/tracing/tracing.go

package tracing

import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/sdk/trace"
)

func InitTracer(serviceName string, endpoint string) (*trace.TracerProvider, error) {
    exporter, err := otlptracegrpc.New(context.Background(),
        otlptracegrpc.WithEndpoint(endpoint),
        otlptracegrpc.WithInsecure(),
    )
    if err != nil {
        return nil, err
    }
    
    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithResource(resource.NewWithAttributes(
            semconv.ServiceNameKey.String(serviceName),
        )),
    )
    
    otel.SetTracerProvider(tp)
    return tp, nil
}
```

- [ ] OpenTelemetry integration
- [ ] Trace spans для evaluation
- [ ] Trace spans для FEEL calls
- [ ] Jaeger/Zipkin export

---

## 📋 Фаза 8: Deployment (2-3 дня)

### 8.1 Цель
Подготовить проект к production deployment.

### 8.2 Задачи

#### 8.2.1 Dockerfile

```dockerfile
# deployments/docker/Dockerfile

# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /dmn-engine ./cmd/server

# Runtime stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /dmn-engine .

EXPOSE 8080

USER nobody

ENTRYPOINT ["./dmn-engine"]
```

- [ ] Multi-stage Dockerfile
- [ ] Minimal runtime image
- [ ] Security best practices

#### 8.2.2 Docker Compose (Development)

```yaml
# deployments/docker/docker-compose.yml

version: '3.8'

services:
  dmn-engine:
    build:
      context: ../..
      dockerfile: deployments/docker/Dockerfile
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://user:pass@postgres:5432/dmn?sslmode=disable
      - FEEL_ENGINE_URL=http://feel-engine:8090
      - REDIS_URL=redis://redis:6379
      - LOG_LEVEL=debug
    depends_on:
      - postgres
      - redis
      - feel-engine

  feel-engine:
    image: camunda/feel-scala-repl:latest
    ports:
      - "8090:8090"

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: pass
      POSTGRES_DB: dmn
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

volumes:
  postgres_data:
```

- [ ] Docker Compose для локальной разработки
- [ ] Health checks
- [ ] Volume persistence

#### 8.2.3 Kubernetes Manifests

```yaml
# deployments/k8s/deployment.yaml

apiVersion: apps/v1
kind: Deployment
metadata:
  name: dmn-engine
spec:
  replicas: 3
  selector:
    matchLabels:
      app: dmn-engine
  template:
    metadata:
      labels:
        app: dmn-engine
    spec:
      containers:
        - name: dmn-engine
          image: your-registry/dmn-engine:latest
          ports:
            - containerPort: 8080
          env:
            - name: DATABASE_URL
              valueFrom:
                secretKeyRef:
                  name: dmn-secrets
                  key: database-url
          resources:
            requests:
              memory: "256Mi"
              cpu: "250m"
            limits:
              memory: "512Mi"
              cpu: "500m"
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 10
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /ready
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 5
```

- [ ] Deployment manifest
- [ ] Service manifest
- [ ] ConfigMap/Secrets
- [ ] HorizontalPodAutoscaler
- [ ] Ingress

#### 8.2.4 Makefile

```makefile
# Makefile

.PHONY: build test run lint docker-build docker-run

# Build
build:
	go build -o bin/dmn-engine ./cmd/server

# Test
test:
	go test -v -race ./...

test-short:
	go test -v -short ./...

test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run
run:
	go run ./cmd/server

# Lint
lint:
	golangci-lint run

# Docker
docker-build:
	docker build -t dmn-engine:latest -f deployments/docker/Dockerfile .

docker-run:
	docker-compose -f deployments/docker/docker-compose.yml up

# Database
db-migrate:
	migrate -path migrations -database "$(DATABASE_URL)" up

db-migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

# Generate
generate:
	go generate ./...
```

- [ ] Makefile с основными командами
- [ ] CI/CD scripts

---

## 📋 Чеклист для запуска MVP

### Минимальный функционал

- [ ] Парсинг DMN 1.3 XML
- [ ] Валидация DMN модели
- [ ] Выполнение Decision Tables (UNIQUE, FIRST)
- [ ] Интеграция с внешним FEEL engine
- [ ] REST API: deploy, evaluate
- [ ] In-memory storage
- [ ] Health check endpoint
- [ ] Docker deployment
- [ ] Базовые тесты

### Nice-to-have для MVP

- [ ] PostgreSQL storage
- [ ] Redis caching
- [ ] Все Hit Policies
- [ ] DRG с под-решениями
- [ ] Prometheus metrics
- [ ] Batch evaluation

---

## 📅 Timeline

| Фаза | Длительность | Результат |
|------|--------------|-----------|
| 0. Исследование | 2-3 дня | Выбран FEEL engine, структура проекта |
| 1. DMN Parser | 3-4 дня | Работающий парсер и валидатор |
| 2. FEEL Client | 2-3 дня | Интеграция с FEEL engine |
| 3. Engine Core | 4-5 дней | Выполнение решений |
| 4. Storage | 2-3 дня | Персистентность |
| 5. REST API | 2-3 дня | API endpoints |
| 6. Testing | 3-4 дня | Тесты и качество |
| 7. Observability | 2-3 дня | Метрики и логи |
| 8. Deployment | 2-3 дня | Docker/K8s |

**Итого MVP: ~4-5 недель**

---

## 🔗 Полезные ресурсы

1. **DMN 1.3 Specification**: https://www.omg.org/spec/DMN/1.3/PDF
2. **FEEL Reference**: Глава 10 в DMN спецификации
3. **feel-scala**: https://github.com/camunda/feel-scala
4. **DSNTK**: https://github.com/dsntk/dsntk-rs
5. **Camunda DMN Docs**: https://docs.camunda.org/manual/latest/reference/dmn/

---

## ❓ Открытые вопросы

1. Какой FEEL engine выбрать? (feel-scala vs DSNTK)
2. Нужна ли обратная совместимость с Camunda API?
3. Требуется ли поддержка Business Knowledge Models?
4. Какой уровень multi-tenancy нужен?
5. Требуется ли аудит выполнений?

