# Настройка Piko для локальной разработки

## Что было сделано

1. **Обновлена конфигурация** в `config/config.yaml`:
   - `server_url` изменен на `http://localhost:8001` (upstream порт)
   - `link_template` изменен на `http://%s.localhost:8000` (proxy порт с поддоменом)

2. **Создан Docker Compose файл** `docker-compose.piko.yml` для запуска piko сервера

3. **Создан скрипт** `start-piko-server.sh` для удобного запуска

## Запуск Piko сервера

### Вариант 1: Через Docker Compose (рекомендуется)

```bash
# Запустить сервер
docker-compose -f docker-compose.yml up -d

# Проверить статус
docker-compose -f docker-compose.yml ps

# Посмотреть логи
docker-compose -f docker-compose.yml logs -f

# Остановить сервер
docker-compose -f docker-compose.yml down
```

### Вариант 2: Через скрипт

```bash
./start-piko-server.sh
```

## Порты

- **8000** - Proxy порт (для входящих HTTP запросов)
- **8001** - Upstream порт (для подключений от клиентов)
- **8002** - Admin порт (метрики и статус API)
- **8003** - Gossip порт (для кластера)

## Проверка работы

После запуска сервера, ваш сервис `runtimedaemonv2` будет подключаться к piko на `http://localhost:8001`.

**Важно:** Piko определяет endpoint по Host header, а не по пути URL!

### Локальная разработка (с localhost)

Endpoint'ы будут доступны по адресу `http://{endpoint_name}.localhost:8000`.

Например, для endpoint с именем "endpointik":
- ✅ Правильно: `http://endpointik.localhost:8000`
- ❌ Неправильно: `http://localhost:8000/endpointik`

### Развертывание на сервере (только IP адрес)

Если у вас нет домена и доступ только по IP адресу, используйте формат `http://{endpoint_name}.{IP_ADDRESS}:8000`.

**Пример конфигурации для сервера с IP `192.168.1.100`:**

В `config/config.yaml`:
```yaml
network:
  piko:
    version: "0.6.0"
    server_url: "http://192.168.1.100:8001"  # IP вашего сервера
    link_template: "http://%s.192.168.1.100:8000"  # Формат с IP
```

Тогда endpoint будет доступен по адресу: `http://endpointik.192.168.1.100:8000`

**Альтернативный вариант - использование файла hosts:**

На клиентской машине добавьте в `/etc/hosts` (Linux/Mac) или `C:\Windows\System32\drivers\etc\hosts` (Windows):
```
192.168.1.100 endpointik.local
```

Тогда можно использовать: `http://endpointik.local:8000`

### Если вы получаете ошибку "missing endpoint id"

Убедитесь, что:
1. Piko сервер запущен
2. Ваше приложение `runtimedaemonv2` запущено и подключено к piko
3. Endpoint зарегистрирован через вызов ChangeMode с mode=Host_Proxy_VM
4. Вы используете правильный формат URL с поддоменом (или IP адресом)

## Остановка

```bash
docker-compose -f docker-compose.yml down
```

