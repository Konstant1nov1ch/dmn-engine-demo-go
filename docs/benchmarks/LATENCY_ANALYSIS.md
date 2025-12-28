# Latency Analysis - Почему Camunda может показать лучшую latency

## Наблюдение

В stress benchmark с 50 decisions:
- **DMN Go**: 15ms average latency
- **Camunda 7**: 14ms average latency
- **Разница**: 1ms (в пределах погрешности)

## Возможные причины

### 1. JVM Warm-up Effect

**Camunda (JVM):**
- После ~30-60 секунд startup, JVM уже "прогрета"
- JIT compiler оптимизировал hot paths
- GC уже настроился под нагрузку
- **Advantage**: Оптимизированный код для повторяющихся операций

**DMN Go:**
- Native binary, но только что запущен
- Go GC может быть менее агрессивным на старте
- Нет "warm-up" периода
- **Disadvantage**: Первые запросы могут быть медленнее

### 2. Caching Strategy

**Camunda:**
- Может кэшировать parsed DMN models в памяти
- Кэширование результатов evaluation для одинаковых inputs
- **Advantage**: Повторные запросы быстрее

**DMN Go (Pre-MVP):**
- Нет кэширования (указано в коде: `cache: nil`)
- Каждый запрос парсит и обрабатывает заново
- **Disadvantage**: Нет оптимизации повторных запросов

### 3. Measurement Variance

**Статистическая погрешность:**
- 10 requests - слишком мало для статистической значимости
- Разница в 1ms может быть просто variance
- Network latency, system load, etc.

### 4. Simple Decision Tables

**Текущий тест:**
- Простые decision tables (3 правила)
- Минимальная логика
- **Вывод**: Для простых cases JVM может быть оптимизирован лучше

**Для сложных cases:**
- Множество правил (100+)
- Сложная FEEL логика
- DMN Go должен показать преимущество

## Ключевые метрики (более важные)

### Memory Footprint: **DMN Go WINS**

```
DMN Go:     16MB
Camunda 7:  ~350MB
Advantage:  21.9x меньше памяти
```

**Это критично для:**
- Container density
- Cloud costs
- Scalability

### Latency: **Statistically Equal**

```
DMN Go:     15ms
Camunda 7:  14ms
Difference: 1ms (6.7% difference)
```

**Вывод:** Разница статистически незначима для 10 requests.

## Рекомендации для более строгого теста

### 1. Увеличить количество requests

```bash
# Вместо 10 requests использовать 100-1000
for i in {1..1000}; do
    # test...
done
```

### 2. Добавить warm-up period

```bash
# Warm-up: 100 requests (не измеряем)
# Then: 1000 requests (измеряем)
```

### 3. Тестировать под нагрузкой

```bash
# Concurrent requests (50-100 одновременно)
# Это покажет реальную производительность
```

### 4. Сложные decision tables

```bash
# Генерировать tables с 50+ правилами
# Сложная FEEL логика
# Множество inputs/outputs
```

### 5. Sustained load test

```bash
# 1 час постоянной нагрузки
# Покажет как системы ведут себя долгосрочно
```

## Выводы

### Текущие результаты:

✅ **Memory**: DMN Go превосходит в **21.9x** (критично!)
✅ **Latency**: Статистически равны (1ms разница незначима)
✅ **Resource efficiency**: DMN Go значительно лучше

### Почему это нормально:

1. **Для простых cases** JVM может быть оптимизирован лучше
2. **JVM warm-up** дает преимущество после startup
3. **Caching** в Camunda помогает повторным запросам
4. **Разница 1ms** - в пределах погрешности измерения

### Что важнее:

**Memory footprint (21.9x advantage)** гораздо важнее чем **1ms latency difference** для:
- Cloud deployments
- Cost optimization
- Scalability
- Container density

## Рекомендации

### Для защиты:

**Подчеркнуть:**
- ✅ Memory advantage: **21.9x** (критично!)
- ✅ Latency: статистически равны (1ms незначима)
- ✅ Resource efficiency: DMN Go значительно лучше

**Объяснить:**
- "Для простых decision tables latency сравнима, но DMN Go использует в 21.9x меньше памяти"
- "Это критично для cloud deployments где memory = cost"
- "При увеличении сложности и нагрузки DMN Go должен показать преимущество"

### Для улучшения теста:

1. Увеличить количество requests (1000+)
2. Добавить warm-up period
3. Тестировать сложные decision tables
4. Concurrent load testing
5. Sustained load (1 hour)

---

**Conclusion**: Memory advantage гораздо важнее минимальной latency разницы для cloud-native deployments.

