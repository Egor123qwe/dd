# ExpirationNotificationService

## How to launch

1. Write your metadata in config file (redis and kafka)
2. Run command `make run`

### Events that this service generates after the TTL expires

- `expired-session` - expired merchant session
```json
{
    "type": "expired-session",
    "meta": {
        "status": "expired",
        "conn": {
            "session_id": "uuid1",
            "type": "session_id"
        }
    },
    "content": {}
}
```
- `expired-client` - expired client
```json
{
    "type": "expired-client",
    "meta": {
        "status": "expired",
        "conn": {
            "user_id": "uuid1",
            "type": "user_id"
        }
    },
    "content": {}
}
```
- `expired-request` - expired rent
```json
{
    "type": "expired-request",
    "meta": {
        "status": "expired",
        "request_id": "uuid1"
    },
    "content": {}
}
```
- `expired-paid-request` - expired the paid hour period
```json
{
  "type": "expired-paid-request",
  "meta": {
    "status": "expired",
    "request_id": "uuid1"
  },
  "content": {}
}
```
