# !!! WORK IN PROGRESS !!! WORK IN PROGRESS !!! WORK IN PROGRESS !!! WORK IN PROGRESS !!! WORK IN PROGRESS !!! WORK IN PROGRESS !!!

## Запуск

1. Применить миграции:
   goose -dir ./ledger/migrations postgres "postgres://postgres:postgres@localhost:5432/cashapp?sslmode=disable" up

2. Запустить приложение:
   go run ./ledger/cmd/ledger

Для отката последней миграции:
goose -dir ./ledger/migrations postgres "postgres://postgres:postgres@localhost:5432/cashapp?sslmode=disable" down


# !!! WORK IN PROGRESS !!! WORK IN PROGRESS !!! WORK IN PROGRESS !!! WORK IN PROGRESS !!! WORK IN PROGRESS !!! WORK IN PROGRESS !!!
