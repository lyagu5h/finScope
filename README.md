## Запуск

1. Применить миграции:
   goose -dir ./ledger/migrations postgres "$DATABASE_URL" up

2. Запустить приложение (Ledger инициализируется при старте):
   go run ./gateway/cmd/gateway

Ledger при старте подключается к PostgreSQL и Redis,
после чего Gateway начинает принимать HTTP-запросы.

Для отката последней миграции:
goose -dir ./ledger/migrations postgres "$DATABASE_URL" down

Пример проверки кеша отчёта:

curl "http://localhost:8080/api/reports/summary?from=2025-01-01&to=2025-12-31"
curl "http://localhost:8080/api/reports/summary?from=2025-01-01&to=2025-12-31"

Повторный запрос будет обслужен из Redis.