#!/bin/bash
# Скрипт для запуска piko сервера локально

# Проверяем наличие Docker
if command -v docker &> /dev/null; then
    echo "Запускаем piko сервер через Docker..."
    docker-compose -f docker-compose.yml up -d
    echo "Piko сервер запущен. Проверьте статус: docker-compose -f docker-compose.piko.yml ps"
    echo "Остановить: docker-compose -f docker-compose.piko.yml down"
    exit 0
fi

# Проверяем, установлен ли piko-server
if ! command -v piko-server &> /dev/null; then
    echo "piko-server не найден. Устанавливаем..."
    
    # Пытаемся установить через go install
    go install github.com/andydunstall/piko/server@v0.6.0 || {
        echo "Не удалось установить через go install."
        echo "Попробуйте один из вариантов:"
        echo "1. Используйте Docker: docker-compose -f docker-compose.piko.yml up -d"
        echo "2. Скачайте бинарник с https://github.com/andydunstall/piko/releases"
        echo "3. Соберите из исходников: git clone https://github.com/andydunstall/piko && cd piko && make build"
        exit 1
    }
fi

# Запускаем piko сервер
echo "Запускаем piko сервер на localhost:8001 (upstream) и localhost:8000 (proxy)..."
piko-server \
    --upstream.bind-addr=0.0.0.0:8001 \
    --proxy.bind-addr=0.0.0.0:8000 \
    --log.level=info
