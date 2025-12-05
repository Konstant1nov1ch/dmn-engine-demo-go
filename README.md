# DMN Engine Go (Pre-MVP)

Высокопроизводительный DMN-движок на Go для микросервисных архитектур.

**Статус: Pre-MVP** — базовый механизм Decision Definition (хранение и управление).

## Функциональность Pre-MVP

✅ **Реализовано:**
- Парсинг DMN 1.3 XML
- Валидация DMN-моделей
- REST API для управления definitions
- PostgreSQL хранилище
- Версионирование definitions
- Multi-tenancy поддержка

🚧 **В разработке:**
- FEEL expressions evaluation
- Decision Table execution
- Redis caching
- Metrics & tracing

## Quick Start

### 1. Запуск PostgreSQL

```bash
# Запустить PostgreSQL в Docker
make db-up

# Или вручную
docker-compose up -d postgres
```

### 2. Запуск сервера

```bash
# Установить зависимости
go mod download

# Запустить сервер
make run

# Или с debug логами
make run-debug
```

Сервер запустится на http://localhost:8080

### 3. Проверка работы

```bash
# Health check
curl http://localhost:8080/health

# Info
curl http://localhost:8080/api/v1/info
```

## API Reference

### Deploy Definition

```bash
# Из файла
curl -X POST http://localhost:8080/api/v1/definitions \
  -H "Content-Type: application/xml" \
  --data-binary @testdata/dmn/simple_decision.dmn

# Multipart form
curl -X POST http://localhost:8080/api/v1/definitions \
  -F "file=@testdata/dmn/simple_decision.dmn" \
  -F "name=My Decision"

# JSON body
curl -X POST http://localhost:8080/api/v1/definitions \
  -H "Content-Type: application/json" \
  -d '{"name": "Test", "xml": "<?xml version=\"1.0\"?>..."}'
```

### List Definitions

```bash
curl http://localhost:8080/api/v1/definitions
```

### Get Definition

```bash
# Latest version
curl http://localhost:8080/api/v1/definitions/myDecision

# Specific version
curl "http://localhost:8080/api/v1/definitions/myDecision?version=1"
```

### Get Definition XML

```bash
curl http://localhost:8080/api/v1/definitions/myDecision/xml
```

### Get Parsed Model

```bash
curl http://localhost:8080/api/v1/definitions/myDecision/parsed
```

### Get All Versions

```bash
curl http://localhost:8080/api/v1/definitions/myDecision/versions
```

### Delete Definition

```bash
curl -X DELETE http://localhost:8080/api/v1/definitions/myDecision
```

### Multi-tenancy

```bash
# Все операции поддерживают tenant ID через header
curl http://localhost:8080/api/v1/definitions \
  -H "X-Tenant-ID: tenant-123"
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `DATABASE_URL` | - | Full PostgreSQL connection string |
| `DB_HOST` | `localhost` | Database host |
| `DB_PORT` | `5432` | Database port |
| `DB_USER` | `dmn` | Database user |
| `DB_PASSWORD` | `dmn` | Database password |
| `DB_NAME` | `dmn` | Database name |
| `DB_SSLMODE` | `disable` | SSL mode |
| `LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |

## Project Structure

```
dmn-engine-go/
├── cmd/server/main.go       # Entry point
├── internal/
│   ├── api/                 # REST API handlers
│   ├── config/              # Configuration
│   ├── dmn/                 # DMN parser & validator
│   └── storage/             # PostgreSQL repository
├── testdata/dmn/            # Sample DMN files
├── docker-compose.yml       # PostgreSQL setup
├── Makefile                 # Build commands
└── README.md
```

## Make Commands

```bash
make help          # Show all commands
make build         # Build binary
make run           # Run server
make run-debug     # Run with debug logs
make test          # Run tests
make db-up         # Start PostgreSQL
make db-down       # Stop PostgreSQL
make db-psql       # Connect to database
make demo          # Deploy sample & test
```

## Example DMN

```xml
<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20191111/MODEL/"
             id="definitions" name="Example">
    <decision id="myDecision" name="My Decision">
        <decisionTable hitPolicy="UNIQUE">
            <input id="input1">
                <inputExpression typeRef="number">
                    <text>age</text>
                </inputExpression>
            </input>
            <output id="output1" name="result" typeRef="string"/>
            <rule id="rule1">
                <inputEntry><text>&gt;= 18</text></inputEntry>
                <outputEntry><text>"adult"</text></outputEntry>
            </rule>
        </decisionTable>
    </decision>
</definitions>
```

## License

MIT
