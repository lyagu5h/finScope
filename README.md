## Запуск

1. Применить миграции:
   goose -dir ./ledger/migrations postgres "$DATABASE_URL" up

2. Запустить приложение:
   go run ./gateway/cmd/gateway

Для отката последней миграции:
goose -dir ./ledger/migrations postgres "$DATABASE_URL" down
