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


## Cheat mode:

1. Set cheat mode param to "true"
2. Set your system info in runtime.conf



   