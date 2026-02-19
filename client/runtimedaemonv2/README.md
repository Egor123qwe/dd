# Runtime daemon

## Endpoints
You can view proto files in ```/internal/handler/grpc/proto```

## Build Instructions
   To build the application:
   ```bash
   make build
   ```
   then you can start application
   
# Request examples (see responses in postman)

- `Download` - download template
```json
{
  "download": {
    "imageName": "nmatsui/hello-world-api",
    "imageTag": "latest"
  },
  "template": {
    "configuration": {
      "envs": [

      ],
      "ports": [
        {
          "port": "3000",
          "title": "hallo",
          "protocol": "HTTP",
          "auth_available": true
        }
      ],
      "volumes": [
        
      ],
      "useGPU": false
    },
    "data": "...",
    "id": "1",
    "type": "local",
    "version": "2.0"
  }
}
```

- `Get` - get template
- here used filter by template type
```json
{
  "types": [
    "local"
  ]
}
```

- `Remove` - remove template
```json
{
  "templateID": "1"
}
```

- `CleanCache` - clean volumes of template
```json
{
  "templateID": "1",
  "clean_host_usage_cache": true,
  "clean_local_usage_cache": true
}
```

- `GetInfo` - get debug info
```json
{}
```

- `GetLogsStream` - get logs from current template
```json
{
  "offset": 0,
  "rows_in_resp": 1
}
```

- `GetStateStream` - get logs from current template
```json
{}
```

- `ExeccuteEvent` - execute or add event
```json
{
    "event": "StopSharing"
}
```

- `ChangeMode (Disabled)`
```json
{
  "mode": "Disable"
}
```

- `ChangeMode (Local)`
```json
{
  "client_id": "yagor",
  "docker": {
    "auth": {
      "credentials": {
        "login": "yagor",
        "password": "1111"
      },
      "enabled": true
    },
    "container": {
      "templateID": "1"
    },
    "client_user_id": "uuid of this user"
  },
  "mode": "Local"
}
```

- `ChangeMode (Client_P2P_VM)`
```json
{
  "client_id": "yagor",
  "mode": "Client_P2P_VM",
  "network": {
    "tailscale": {
      "auth_key": "tskey-auth-kYqsEBMCWB21CNTRL-9FvxFjuFkMLNCKzTZYv2MLS7UgwofP95",
      "client_id": "uuid"
    }
  }
}
```

- `ChangeMode (Host_P2P_VM)`
- pass in tailscale "auth_key" your own key
```json
{
  "client_id": "yagor",
  "docker": {
    "auth": {
      "credentials": {
        "login": "yagor",
        "password": "1111"
      },
      "enabled": true
    },
    "container": {
      "templateID": "1"
    },
    "client_user_id": "uuid of user, that will use this container"
  },
  "network": {
    "tailscale": {
      "auth_key": "tskey-auth-kYqsEBMCWB21CNTRL-9FvxFjuFkMLNCKzTZYv2MLS7UgwofP95",
      "client_id": "uuid"
    }
  },
  "mode": "Host_P2P_VM"
}
```

- `ChangeMode (Host_Proxy_VM)`
```json
{
  "client_id": "yagor",
  "docker": {
    "auth": {
      "credentials": {
        "login": "yagor",
        "password": "1111"
      },
      "enabled": true
    },
    "container": {
      "templateID": "1"
    },
    "client_user_id": "uuid of user, that will use this container"
  },
  "network": {
    "piko": {
      "auth_key": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.XbPfbIHMI6arZ3Y922BhjWgQzWXcXNrz0ogtVhfEd2o",
      "endpoints": [
        {
          "template_port": "3000",
          "name": "endpointik"
        }
      ]
    }
  },
  "mode": "Host_Proxy_VM"
}
```

# For test how work with volumes, you can use postgres

- `Download` - download template
```json
{
  "download": {
    "imageName": "postgres",
    "imageTag": "latest"
  },
  "template": {
    "configuration": {
      "envs": [
            "-e", "POSTGRES_USER=user",
            "-e", "POSTGRES_PASSWORD=pass",
            "-e", "POSTGRES_DB=db" 
      ],
      "ports": [
        {
          "port": "5432",
          "title": "admin",
          "protocol": "HTTP",
          "auth_available": false
        }
      ],
      "volumes": [
        "/var/lib/postgresql/data"
      ],
      "useGPU": false
    },
    "data": "...",
    "id": "2",
    "type": "Local",
    "version": "2.0"
  }
}
```