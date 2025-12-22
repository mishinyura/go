## Users API (Gin)

Минималистичный микросервис на Go:
- CRUD по пользователям (Gin + чистая архитектура, in-memory хранилище).
- Пример интеграции с Google Sheets через сервисный аккаунт.

### Структура

```
cmd/
  server/
    main.go           # точка входа
internal/
  handler/           # HTTP-слой (Gin)
  service/           # бизнес-логика
  repository/        # репозиторий (in-memory)
  model/             # доменные модели и DTO
  docs/              # минимальные Swagger-доки (можно перегенерировать swag CLI)
```

### Установка и запуск

Установка зависимостей и генерация Swagger (опционально):

```bash
make deps tidy
make swag
```

Запуск сервера:

```bash
make run
# сервер слушает :8080
```

Документация:

- Swagger UI: `http://localhost:8080/swagger/index.html`
- ReDoc: `http://localhost:8080/redoc`

### User CRUD

- GET    `/api/v1/users` — список
- POST   `/api/v1/users` — создать
- GET    `/api/v1/users/{id}` — получить по ID
- PATCH  `/api/v1/users/{id}` — частичное обновление
- DELETE `/api/v1/users/{id}` — удалить

Примеры запросов:

```bash
# создать
curl -s -X POST http://localhost:8080/api/v1/users \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","email":"alice@example.com"}'

# список
curl -s http://localhost:8080/api/v1/users | jq .

# получить по id
curl -s http://localhost:8080/api/v1/users/<id>

# частичное обновление
curl -s -X PATCH http://localhost:8080/api/v1/users/<id> \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice2"}'

# удалить
curl -s -X DELETE http://localhost:8080/api/v1/users/<id> -i
```

- Swagger аннотации находятся в `internal/handler/*.go`.
- Перегенерация OpenAPI: `swag init -g cmd/server/main.go -o internal/docs`.

---

## Google Sheets Demo

### Что реализовано

- `POST /api/v1/demo/export` — записывает в таблицу произвольный набор строк (приходит в теле запроса).
- `GET /api/v1/demo/import` — читает указанную таблицу (`spreadsheet_id`, `sheet_name` в query).
- `GET /api/v1/demo/download` — читает «таблицу по умолчанию», параметры берёт из переменных окружения.

Бизнес-логика лежит в `internal/service/simple_sheet_service.go`, низкоуровневый клиент Google — в `internal/service/sheets_client.go`.

### Подготовка Google Sheets

1. Создать проект в Google Cloud Console и включить **Google Sheets API**.
2. Создать сервисный аккаунт (IAM & Admin → Service Accounts), назначить роль `Editor` или `Owner`, скачать JSON-ключ.
3. Создать таблицу в Google Sheets, поделиться ею с сервисным аккаунтом (Editor).
4. Настроить переменные окружения:
   ```bash
   export GOOGLE_APPLICATION_CREDENTIALS=/abs/path/to/key.json
   export DEMO_SPREADSHEET_ID=<ID таблицы>
   export DEMO_SHEET_NAME=Report
   ```
5. Запустить сервер (`go run cmd/server/main.go` или `make run`) и открыть Swagger.

### Примеры curl

```bash
# Записать произвольные данные (clear=true очищает диапазон A/B)
curl -s -X POST http://127.0.0.1:8080/api/v1/demo/export \
  -H 'Content-Type: application/json' \
  -d '{
        "spreadsheet_id": "Ваш ID таблицы",
        "sheet_name": "Report",
        "clear": true,
        "rows": [
          {"category": "Кофе", "amount": 390},
          {"category": "Коммуналка", "amount": 8200},
          {"category": "Такси", "amount": 670}
        ]
      }'

# Прочитать таблицу, указанную в env (Download)
curl -s http://127.0.0.1:8080/api/v1/demo/download

# Прочитать любую таблицу по параметрам
curl -s "http://127.0.0.1:8080/api/v1/demo/import?spreadsheet_id=Ваш_ID_Таблицы_name=Report"
```