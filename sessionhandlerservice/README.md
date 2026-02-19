# SessionHandlerService

### This server handle start/stop flow, also store some ids in redis for other services, Also handle ttl events

## How to launch

1. Write your metadata in config file 
2. Run command `make run`

## START FLOW:
 ### 1. (from client):
- `start-session` - request to start session
- `mode` - must be ["proxy", "p2p"] proxy - piko, p2p - tailscale
```json
{
    "type": "start-session",
    "meta": {
        "conn": {
            "user_id": "uuid",
            "conn_id": "uuid",
            "type": "all_ids"
        }
    },
    "content": {
        "session_id": "uuid",
        "settings": {
            "mode": "proxy",
            "template_id": "uuid"
        }
    }
}
```

### 2. (to merchant from server):
- `merchant-start-rent` - request to start rent
- `mode` - must be ["proxy", "p2p"] proxy - piko, p2p - tailscale
#### IMPORTANT: in network you will get piko or tailscale (one of those)
```json
{
    "type": "merchant-start-rent",
    "meta": {
        "status": "ok",
        "conn": {
            "conn_id": "uuid",
            "type": "conn_id"
        }
    },
    "content": {
        "request_id": "uuid",
        "session_id": "uuid",
        "settings": {
              "mode": "proxy",
              "template": {
                    "id": "uuid",
                    "image_name": "name",
                    "image_tag": "tag",
                    "version": "v0.0.0",
                    "authentication": {
                          "login": "login",
                          "password": "password"
                    }
              },
              "network": {
                    "piko": {
                        "auth_key": "key",
                        "endpoints": {
                            "template_port": 777,
                            "name": "endpoint"
                        }
                    },
                    "tailscale": {
                        "auth_key": "auth_key"
                    }
              }
        }
    }
}
```

### 3. (to client from server):
- `session-status-updated` - request to start rent
- `status` - must be ["running", "pending" "stopped"]
```json
{
    "type": "session-status-updated",
    "meta": {
        "status": "ok",
        "conn": {
            "user_id": "uuid",
            "type": "user_id"
        }
    },
    "content": {  
        "request_id": "uuid",
        "session_id": "uuid",
        "status": "pending",
        "status_msg": ""
    }
}
```

### 4. (to output topic (kafka) from server):
- `init-rent` - request to start watch request_id in redis
```json
{
    "type": "init-rent",
    "meta": {},
    "content": {  
        "request_id": "uuid"
    }
}
```

#### And also set request_id in redis
#### And also set client_id in redis


## After first handshake, merchant should to start model. When he is ready, starting continue:

### 5. (from merchant):
- `rent-request-status-updated` - request to start rent
- `status` - must be ["running", "error"] 
```json
{
    "type": "rent-request-status-updated",
    "meta": {
        "conn": {
            "user_id": "uuid",
            "type": "user_id"
        }
    },
    "content": {
        "request_id": "uuid",
        "status": "running",
        "status_msg": "..."
    }
}
```

## P.S. If status - error, service start "STOP FLOW" (read below)

### 6. (from server to client):
- `mode` - must be ["proxy", "p2p"] proxy - piko, p2p - tailscale
- `client-start-rent` - request to start rent
```json
{
      "type": "client-start-rent",
      "meta": {
          "status": "ok",
          "conn": {
              "conn_id": "uuid",
              "type": "conn_id"
          }
      },
      "content": {
          "request_id": "uuid",
          "session_id": "uuid",
          "settings": {
                "mode": "proxy",
                "template": {
                      "authentication": {
                            "login": "login",
                            "password": "password"
                      }
                },
                "network": {
                      "piko": {
                          "endpoints": {
                              "title": "title",
                              "type":  "http",
                              "name": "endpoint"
                          }
                      },
                      "tailscale": {
                          "auth_key": "auth_key"
                      }
                }
          }
   
      }
}
```

### 7. (from server to merchant):
- `session-status-updated` - request to start rent
- `status` - must be ["running", "pending" "stopped"]
```json
{
    "type": "session-status-updated",
    "meta": {
        "status": "ok",
        "conn": {
            "conn_id": "uuid",
            "type": "conn_id"
        }
    },
    "content": {
        "request_id": "uuid",
        "session_id": "uuid",
        "status": "running",
        "status_msg": ""
    }
}
```

## STOP FLOW:
### 1. (from client or merchant):
- `stop-session` - request to start session
```json
{
    "type": "stop-session",
    "meta": {
        "status": "ok",
        "conn": {
            "user_id": "uuid",
            "conn_id": "uuid",
            "type": "all_ids"
        }
    },
    "content": {
        "request_id": "uuid",
        "reason": "reason of stop"
    }
}
```

### 2. (from server to merchant and client):
- `session-status-updated` - request to start rent
- `status` - must be ["running", "pending" "stopped"]
```json
{
    "type": "session-status-updated",
    "meta": {
        "status": "ok",
        "conn": {
            "conn_id": "uuid",
            "type": "conn_id"
        }
    },
    "content": {
        "request_id": "uuid",
        "session_id": "uuid",
        "status": "stopped",
        "status_msg": "reason of stop"
    }
}
```

#### And also delete request_id from redis



## ERROR CASE:
```json
{
    "type": "the event that caused the error",
    "meta": {
        "status": "err",
        "conn": {
            "conn_id": "uuid",
            "type": "conn_id"
        },
        "err": {
            "code": "err",
            "msg": "msg"
        }
    },
    "content": {}
}
```

## TTL EXPIRE HANDLING:
### This server handle "expired-request" and "expired-client" events
- `expired-request` - stop one rent
- `expired-client` - stop all client`s rents
