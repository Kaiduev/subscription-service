# Subscription Service

Тестовое задание Golang (сервис для управления подписками пользователей).

## Запуск

Требуется установленный Docker.

```bash
make run
```
или  
```bash
docker compose up --build
```

После старта:
- **Swagger UI**: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)  
- **Healthcheck**: [http://localhost:8080/healthz](http://localhost:8080/healthz)  

Остановить и очистить:
```bash
make down
```
или  
```bash
docker compose down -v
```

## Переменные окружения

```
DB_USER=app
DB_PASSWORD=secret
DB_NAME=subscriptions
DB_HOST=db
DB_PORT=5432
DB_SSLMODE=disable
```

## Быстрый тест API

Создание подписки:
```bash
curl -s -X POST http://localhost:8080/subscriptions/   -H "Content-Type: application/json"   -d '{
    "service_name": "Yandex Plus",
    "price": 400,
    "user_id": "24b54fed-d271-403f-ad0d-453812f937bb",
    "start_date": "2025-07"
  }' | jq
```

Получить подписку по ID:
```bash
curl -s "http://localhost:8080/subscriptions/<id>" | jq
```

Список подписок:
```bash
curl -s "http://localhost:8080/subscriptions?limit=10&offset=0&service_name=Yandex" | jq
```

Обновить подписку:
```bash
curl -s -X PUT http://localhost:8080/subscriptions/<id> \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "Yandex Music",
    "price": 500,
    "user_id": "24b54fed-d271-403f-ad0d-453812f937bb",
    "start_date": "2025-07",
    "end_date": "2025-09"
  }' | jq
```

Удалить подписку:
```bash
curl -s -X DELETE "http://localhost:8080/subscriptions/<id>" | jq
```

Сумма по подпискам за период:
```bash
curl -s "http://localhost:8080/subscriptions/summary?start=2025-07&end=2025-09&user_id=24b54fed-d271-403f-ad0d-453812f937bb" | jq
```