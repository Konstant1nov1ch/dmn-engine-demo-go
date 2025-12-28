# Camunda Platform 7.24 Dockerfile
FROM camunda/camunda-bpm-platform:7.24.0

# Опционально: добавление собственных DMN/BPMN файлов
# COPY *.bpmn /camunda/configuration/resources/
# COPY *.dmn /camunda/configuration/resources/

# Опционально: добавление кастомных плагинов
# COPY plugins/*.jar /camunda/lib/

# Порт Camunda Web Apps
EXPOSE 8080

# Здоровье контейнера
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD curl -f http://localhost:8080/camunda/app/welcome/default/ || exit 1

