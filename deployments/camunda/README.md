# Camunda 7 для бенчмаркинга

## Быстрый старт

### 1. Запустить Camunda 7

```bash
cd deployments/camunda
docker-compose up -d
```

**Это запустит:**
- Camunda 7.20 на порту **8081** (чтобы не конфликтовать с DMN Go на 8080)
- PostgreSQL на порту **5433** (чтобы не конфликтовать с DMN Go postgres на 5432)

### 2. Дождаться полной загрузки

Camunda требует ~30-60 секунд для полной инициализации.

```bash
# Проверить статус
docker-compose logs -f camunda

# Проверить health
curl http://localhost:8081/engine-rest/engine
```

### 3. Открыть Camunda Cockpit

```
URL: http://localhost:8081/camunda/app/cockpit/default/
User: admin
Password: admin
```

### 4. Deploy DMN

**Вариант A: Через UI**
1. Открыть Cockpit
2. Navigate to Deployments
3. Upload `testdata/dmn/simple_decision.dmn`

**Вариант B: Через REST API**
```bash
curl -X POST http://localhost:8081/engine-rest/deployment/create \
  -F "deployment-name=test" \
  -F "deployment-source=local" \
  -F "data=@../../testdata/dmn/simple_decision.dmn"
```

**Вариант C: Auto-deploy (если настроено)**
- Файлы в `./deployments/dmn/` должны автоматически деплоиться

## Тестирование

### Manual test

```bash
# Evaluate decision
curl -X POST http://localhost:8081/engine-rest/decision-definition/key/eligibility/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "variables": {
      "age": {"value": 25, "type": "Integer"}
    }
  }' | jq .
```

### Automated benchmark

```bash
# Вернуться в корень проекта
cd ../..

# Запустить Camunda benchmark
./scripts/test_camunda.sh

# Или сравнить оба
./scripts/compare_both.sh
```

## Остановить

```bash
cd deployments/camunda
docker-compose down

# С удалением данных
docker-compose down -v
```

## Troubleshooting

### Camunda не запускается

```bash
# Проверить логи
docker-compose logs camunda

# Проверить память
docker stats camunda-7
```

### Порт занят

Если порт 8081 занят, измените в `docker-compose.yml`:
```yaml
ports:
  - "8082:8080"  # Используйте другой порт
```

### База данных

Camunda создаст свою схему автоматически при первом запуске.

## Конфигурация

**JVM Settings** в docker-compose.yml:
```yaml
environment:
  - JAVA_OPTS=-Xmx512m -Xms256m
```

- `-Xmx512m` - максимальная heap память
- `-Xms256m` - начальная heap память

Можно изменить для тестирования разных конфигураций.

## Архитектура

```
┌─────────────────────┐
│   Camunda 7         │
│   (JVM)             │
│   Port: 8081        │
└──────────┬──────────┘
           │
           │ JDBC
           │
┌──────────▼──────────┐
│   PostgreSQL        │
│   Port: 5433        │
└─────────────────────┘
```

## Для честного сравнения

Оба engine используют:
- ✅ PostgreSQL для persistence
- ✅ Тот же DMN файл
- ✅ Одинаковое железо (localhost)
- ✅ Изолированные порты (no conflicts)

Различия:
- DMN Go: Native binary, direct execution
- Camunda 7: JVM-based, Docker container

