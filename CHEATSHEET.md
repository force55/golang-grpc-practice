# Crypto Notifier — Шпаргалка команд

## buf (protobuf тулчейн)

```bash
# Перевірити proto файли на помилки (лінтинг)
buf lint proto

# Згенерувати Go код з proto файлів (результат у gen/)
buf generate proto

# Перевірити чи нові зміни в proto не ламають зворотну сумісність
buf breaking proto --against '.git#subdir=proto'
```

## go-task (task runner)

```bash
# Показати всі доступні таски
task --list

# Запустити конкретний таск
task <назва>
```

## Docker Compose

```bash
# Піднати всі сервіси (postgres і т.д.)
docker compose up -d

# Зупинити
docker compose down

# Подивитись логи
docker compose logs -f postgres
```

## Go

```bash
# Встановити залежності
go mod tidy

# Запустити сервер
go run ./cmd/server

# Запустити тести
go test ./...

# Запустити тести з verbose output
go test -v ./...
```

## Тестування Connect-RPC ендпоінтів

```bash
# Unary RPC через curl (Connect protocol — звичайний HTTP POST з JSON)
curl -X POST http://localhost:8081/price.v1.PriceService/GetPrice \
  -H "Content-Type: application/json" \
  -d '{"symbol": "BTCUSDT"}'

# Server streaming (curl не підходить — потрібен buf curl)
buf curl --protocol connect \
  http://localhost:8081/price.v1.PriceService/Subscribe \
  --schema proto \
  -d '{"symbols": ["BTCUSDT"]}'

# Створити алерт
curl -X POST http://localhost:8081/alert.v1.AlertService/CreateAlert \
  -H "Content-Type: application/json" \
  -d '{"userId": "<UUID>", "symbol": "BTCUSDT", "targetPrice": 60000, "direction": "below"}'

# Список алертів юзера
curl -X POST http://localhost:8081/alert.v1.AlertService/ListAlerts \
  -H "Content-Type: application/json" \
  -d '{"userId": "<UUID>"}'

# Видалити алерт
curl -X POST http://localhost:8081/alert.v1.AlertService/DeleteAlert \
  -H "Content-Type: application/json" \
  -d '{"id": "<ALERT_UUID>"}'
```

## PostgreSQL (seed data)

```bash
# Створити тестового юзера
docker exec -it crypto-notifier-db-1 psql -U postgres -c \
  "INSERT INTO users (name) VALUES ('testuser') RETURNING id;"

# Подивитись всіх юзерів
docker exec -it crypto-notifier-db-1 psql -U postgres -c "SELECT * FROM users;"

# Подивитись всі алерти
docker exec -it crypto-notifier-db-1 psql -U postgres -c "SELECT * FROM alerts;"
```

## Git

```bash
git status
git add <files>
git commit -m "message"
```

## Міграції (golang-migrate)

```bash
# Connection string формат:
# postgresql://USER:PASSWORD@HOST:PORT/DATABASE?sslmode=disable
# Значення беруться з docker-compose.yml (POSTGRES_USER, POSTGRES_PASSWORD, порт, POSTGRES_DB)

# Застосувати всі міграції
migrate -path migrations -database "postgresql://postgres:postgres@localhost:5433/postgres?sslmode=disable" up

# Відкатити останню міграцію
migrate -path migrations -database "postgresql://postgres:postgres@localhost:5433/postgres?sslmode=disable" down 1

# Подивитись поточну версію міграцій
migrate -path migrations -database "postgresql://postgres:postgres@localhost:5433/postgres?sslmode=disable" version

# Створити нову міграцію
migrate create -ext sql -dir migrations -seq назва_міграції
```

## sqlc

```bash
# Згенерувати Go код з SQL запитів
sqlc generate
```