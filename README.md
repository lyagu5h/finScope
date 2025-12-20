# !!! WORK IN PROGRESS !!! WORK IN PROGRESS !!! WORK IN PROGRESS !!! WORK IN PROGRESS !!! WORK IN PROGRESS !!! WORK IN PROGRESS !!!

## Запуск

1. Применить миграции:
   goose -dir ./ledger/migrations postgres "$DATABASE_URL" up

2. Запустить приложение:
   go run ./ledger/cmd/ledger

Для отката последней миграции:
goose -dir ./ledger/migrations postgres "$DATABASE_URL" down


# !!! WORK IN PROGRESS !!! WORK IN PROGRESS !!! WORK IN PROGRESS !!! WORK IN PROGRESS !!! WORK IN PROGRESS !!! WORK IN PROGRESS !!!
