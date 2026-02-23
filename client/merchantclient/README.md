# MerchantClient

## Build: 
   ```bash
   make build
   ```

## Usage:
    You can use flags, envs or defaul values

   ```bash
    ./merchantclient --token=auth_token
   ```

    see more information about flags in ./merchantclient --help

## Params rules

1. You also can use envs, but if you set value in envs and flags, flags will chosen first
2. If envs and flags not set, app will try to use default value

## Params table:

| Logical param       | Default values                  | Envs keys           | Flags            |
|---------------------|---------------------------------|---------------------|------------------|
| auth key            | ""                              | ROY9_AUTH_TOKEN     | --token          |
| state machine url   | wss://sm-prod.roy9.ru/api/v1/ws | ROY9_CONNECTION_URL | --connection-url |
| runtime daemon port | 30099                           | ROY9_RD_PORT        | --rd-port        |
| logger level        | INFO                            | ROY9_CLIENT_VERBOSE | --verbose        |
| cheat mode          | false                           | CHEAT_MODE          | --cheat-mode     |
| backend host (IP)   | —                                | ROY9_BACKEND_HOST   | —                |


## Запуск на другом ПК (бэкенд на ноутбуке в сети)

Скопируйте `.env.example` в `.env` и укажите IP ноутбука с бэкендом:

```bash
ROY9_BACKEND_HOST=192.168.1.100
```

Тогда в `auth.service_url`, `status_check.url` и `state_machine.connection_url` вместо `localhost` будет подставлен этот хост. На ноутбуке должен быть поднят compose с `docker-compose.lan.yml` (порты 8052, 8090 доступны по IP).

## Cheat mode:

1. Set cheat mode param to "true"
2. Set your system info in runtime.conf



   