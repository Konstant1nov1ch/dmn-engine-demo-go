# 🔬 Real Benchmark Guide - Side-by-Side Comparison

## Цель

Провести **честное, воспроизводимое** сравнение DMN Engine Go с Camunda 7 на одном и том же железе.

### 🔬 Методология: Sequential Testing

Системы тестируются **по очереди**, а не одновременно, чтобы избежать:
- ❌ Конкуренции за CPU
- ❌ Конкуренции за память  
- ❌ I/O contention
- ❌ Влияния друг на друга

**Порядок:**
1. Тест DMN Go (единственная система)
2. Остановка DMN Go
3. Запуск и тест Camunda (единственная система)
4. Сравнение результатов

## Подготовка (2 минуты)

### Автоматический запуск (Recommended)

Скрипт `compare_both.sh` автоматически:
- ✅ Запустит и протестирует DMN Go
- ✅ Остановит DMN Go
- ✅ Запустит и протестирует Camunda
- ✅ Сравнит результаты

**Всё что нужно:**
```bash
make db-up  # Только база данных
make benchmark-compare  # Всё остальное автоматически
```

### Ручной запуск (Optional)

Если хочешь сам контролировать процесс:

**1. Подготовить базу:**
```bash
make db-up
```

**2. Собрать DMN Go:**
```bash
make build
```

**3. Запустить тесты:**
```bash
./scripts/compare_both.sh
```

Скрипт сам запустит и остановит каждую систему по очереди.

## Тестирование (10-15 минут)

### Option 1: Автоматическое сравнение (Recommended)

```bash
./scripts/compare_both.sh
```

Этот скрипт автоматически:
- ✅ Измерит оба engine
- ✅ Сравнит результаты
- ✅ Сгенерирует отчёт

### Option 2: Ручное тестирование

#### Test DMN Engine Go

```bash
# Memory
ps aux | grep dmn-engine

# Latency (50 requests)
for i in {1..50}; do
  curl -w "%{time_total}\n" -o /dev/null -s \
    -X POST http://localhost:8080/api/v1/evaluate \
    -H "Content-Type: application/json" \
    -d '{"decisionKey":"eligibility","variables":{"age":25}}'
done | awk '{sum+=$1; count++} END {print "Avg:", sum/count*1000, "ms"}'

# Throughput
ab -n 1000 -c 50 -p /tmp/request.json -T 'application/json' \
  http://localhost:8080/api/v1/evaluate
```

#### Test Camunda 7

```bash
# Memory
docker stats camunda-7 --no-stream

# Latency (50 requests)
for i in {1..50}; do
  curl -w "%{time_total}\n" -o /dev/null -s \
    -X POST http://localhost:8081/engine-rest/decision-definition/key/eligibility/evaluate \
    -H "Content-Type: application/json" \
    -d '{"variables":{"age":{"value":25,"type":"Integer"}}}'
done | awk '{sum+=$1; count++} END {print "Avg:", sum/count*1000, "ms"}'

# Throughput
cat > /tmp/camunda_request.json <<EOF
{"variables":{"age":{"value":25,"type":"Integer"}}}
EOF

ab -n 1000 -c 50 -p /tmp/camunda_request.json -T 'application/json' \
  http://localhost:8081/engine-rest/decision-definition/key/eligibility/evaluate
```

## Метрики для сравнения

### 1. Cold Start Time

**DMN Go:**
```bash
pkill -f dmn-engine
time make run  # Measure until ready
```

**Camunda 7:**
```bash
cd deployments/camunda
docker-compose restart camunda
docker-compose logs -f camunda  # Note startup time
```

### 2. Memory Footprint

**DMN Go:**
```bash
ps -o rss= -p $(pgrep -f dmn-engine) | awk '{print $1/1024 "MB"}'
```

**Camunda 7:**
```bash
docker stats camunda-7 --no-stream --format "{{.MemUsage}}"
```

### 3. Latency (P50, P90, P99)

Используйте Apache Bench или `scripts/load_test.py`:

```bash
# DMN Go
python3 scripts/load_test.py \
  --url http://localhost:8080 \
  --key eligibility \
  --users 50 --requests 20

# Camunda 7
# (Нужно адаптировать load_test.py для Camunda API)
```

### 4. Throughput (req/s)

Apache Bench покажет "Requests per second".

### 5. Container Density (theoretical)

Рассчитать на основе memory footprint:
```
8GB server / memory_per_instance = instances
```

## Ожидаемые результаты

### Hypothesis

| Metric | DMN Go | Camunda 7 | Expected Advantage |
|--------|--------|-----------|-------------------|
| Cold Start | ~100ms | ~3000ms | ~30x faster |
| Memory | ~40MB | ~250MB | ~6x less |
| Latency | ~5-10ms | ~20-40ms | ~3-5x faster |
| Throughput | ~2000/s | ~700/s | ~3x more |

**Note:** Реальные результаты будут зависеть от:
- Hardware (CPU, RAM)
- Network latency (localhost vs remote)
- System load
- Docker overhead

## Сохранение результатов

### 1. Автоматический отчёт

`compare_both.sh` создаст `comparison_results.txt`.

### 2. Ручной отчёт

Создайте файл `real_benchmark_results.md`:

```markdown
# Real Benchmark Results

Date: [DATE]
Environment: [OS / CPU / RAM]

## Results

| Metric | DMN Go | Camunda 7 | Advantage |
|--------|--------|-----------|-----------|
| Cold Start | [X]ms | [Y]ms | [Y/X]x |
| Memory | [X]MB | [Y]MB | [Y/X]x |
| Avg Latency | [X]ms | [Y]ms | [Y/X]x |
| P99 Latency | [X]ms | [Y]ms | [Y/X]x |
| Throughput | [X]/s | [Y]/s | [X/Y]x |

## Conclusion

[Your analysis]
```

## Обновление документации

После **реальных** замеров:

1. **Обновить METHODOLOGY.md:**
   - Изменить статус с "documented" на "measured"
   - Добавить реальные результаты
   - Описать test environment

2. **Обновить JUSTIFICATION.md:**
   - Заменить estimates на actual measurements
   - Добавить disclaimer о test environment

3. **Обновить все benchmark docs:**
   - BENCHMARK_RESULTS.md
   - BENCHMARK_SUMMARY.md
   - README.md

4. **Обновить quick_compare.sh:**
   - Использовать реальные measured values

## Troubleshooting

### Camunda слишком медленный

**Возможные причины:**
1. Docker overhead
2. Недостаточно выделено памяти JVM
3. Database not optimized

**Решения:**
- Увеличить JVM heap: `-Xmx1024m`
- Запустить Camunda не в Docker (для fair comparison)
- Оптимизировать PostgreSQL

### DMN Go слишком быстрый (?)

**Проверить:**
- Реально ли выполняются запросы?
- Database hit происходит?
- Правильно ли измеряется time?

**Честно документировать:**
Если результаты слишком хороши, проверить и перепроверить.

## Best Practices

### 1. Multiple Runs

Запустите тесты несколько раз и усредните:

```bash
for i in {1..5}; do
  ./scripts/compare_both.sh
  sleep 10
done
```

### 2. Warm-up

Перед измерением latency/throughput:

```bash
# Warm-up (100 requests)
ab -n 100 -c 10 -p request.json -T 'application/json' [URL]

# Then measure
ab -n 1000 -c 50 -p request.json -T 'application/json' [URL]
```

### 3. Same Conditions

- Закрыть другие приложения
- Стабильная сеть
- Одинаковая нагрузка
- Одно и то же время суток

### 4. Document Everything

- OS version
- CPU model
- RAM size
- Docker version
- Go version
- Java version (в Camunda container)

## Academic Integrity

**После реальных бенчмарков:**

✅ **Честно:**
- Описать test setup
- Показать raw data
- Объяснить anomalies (если есть)
- Acknowledge limitations

✅ **Reproducible:**
- Скрипты предоставлены
- Environment documented
- Шаги описаны

✅ **Fair:**
- Одинаковое железо
- Одинаковые условия
- Честное сравнение

---

**Ready to start?**

```bash
# 1. Start both engines
make db-up && make run
cd deployments/camunda && docker-compose up -d && cd ../..

# 2. Run comparison
./scripts/compare_both.sh

# 3. Review results
cat comparison_results.txt
```

**Let's measure! 🔬**

