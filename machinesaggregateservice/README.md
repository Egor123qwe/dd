# MachinesAggregateService

make run - сборка и запуск бинарника в docker

make migrate - миграции с фикстурами для теста

http://localhost:8000/api/v1 - base url

swagger - api/v1/swagger/index.html

api/v1/exchange/category/easy-plan [get] - получение облегченного тарифа 

```json 
[
  {
    "id": "string",
    "name": "string",
    "price": 0,
    "total_sessions": 0
    "sessions" : [
      "string"
    ]
  },
]
```
api/v1/exchange/category/gpu [get] - метод для получения названий видеокарт из сущности gpu_dict с количеством соответствующих сессий и минимальной ценой

```json
{
"category": [
    {
      "id": "string",
      "name": "string",
      "price_min": 0,
      "total_sessions": 0,
      "totalvram": "string"
    }
  ]
}
```
api/v1/exchange/category/gpu/{gpu_name_id}/session [get] - метод для получения сессий по id сущности gpu_dict(название определенной видеокарты)  

```json
{
    "category": {
    "id": "string",
    "name": "string",
    "price_min": 0,
    "total_sessions": 0,
    "totalvram": "string"
  },
  "max_ram": 0,
  "max_storage": 0,
  "min_ram": 0,
  "min_storage": 0,
  "sessions": [
    {
      "available_ram": 0,
      "cpus": [
        {
          "available": 0,
          "id": "string",
          "name": "string",
          "price": 0,
          "total": 0
        }
      ],
      "created_at": "string",
      "deleted_at": "string",
      "gpus": [
        {
          "availablevram": 0,
          "avg_dlperf": 0,
          "id": "string",
          "name": "string",
          "price": 0,
          "totalvram": 0,
          "usedvram": 0
        }
      ],
      "load_speed": 0,
      "ping": 0,
      "prepull": [
        {
          "id": "string",
          "version": "string"
        }
      ],
      "price_internet": 0,
      "price_ram": 0,
      "session_id": "string",
      "storage": [
        {
          "available": 0,
          "bandwidth": 0,
          "id": "string",
          "name": "string",
          "price": 0,
          "total": 0,
          "type": "string",
          "used": 0
        }
      ],
      "total_price": 0,
      "total_ram": 0,
      "upload_speed": 0,
      "used_ram": 0
    }
  ]
}
```
api/v1/hardware/gpu/ [get] - получение реальных видеокарт с РЕАЛЬНЫМ названием прописанным в сессии
```json 
{
      "gpus": [
    {
      "availablevram": 0,
      "avg_dlperf": 0,
      "id": "string",
      "name": "string",
      "price": 0,
      "totalvram": 0,
      "usedvram": 0
    }
  ]
}
```

api/v1/hardware/gpu/{gpu_id} - получение реальной видеокарты, соотвествующей определенной сессии


```json
{
    "gpus": [
    {
      "availablevram": 0,
      "avg_dlperf": 0,
      "id": "string",
      "name": "string",
      "price": 0,
      "totalvram": 0,
      "usedvram": 0
    }
  ]
}
```

api/v1/session/{session_id} - получение сессии по ее id

```json 
{
    {
  "category": {
    "gpu": {
          "id": "string",
          "name ": "string",
          "totalvram": "string"
        }
  },
  "session": {
    "available_ram": 0,
    "cpus": [
      {
        "available": 0,
        "id": "string",
        "name": "string",
        "price": 0,
        "total": 0
      }
    ],
    "created_at": "string",
    "deleted_at": "string",
    "gpus": [
      {
        "availablevram": 0,
        "avg_dlperf": 0,
        "id": "string",
        "name": "string",
        "price": 0,
        "totalvram": 0,
        "usedvram": 0
      }
    ],
    "load_speed": 0,
    "ping": 0,
    "prepull": [
      {
        "id": "string",
        "version": "string"
      }
    ],
    "price_internet": 0,
    "price_ram": 0,
    "session_id": "string",
    "storage": [
      {
        "available": 0,
        "bandwidth": 0,
        "id": "string",
        "name": "string",
        "price": 0,
        "total": 0,
        "type": "string",
        "used": 0
      }
    ],
    "total_price": 0,
    "total_ram": 0,
    "upload_speed": 0,
    "used_ram": 0
  }
}
}
```