build:
	go build ./...

test:
	go test ./...

proto:
	make proto

migrate-up:
	goose -dir ./ledger/migrations postgres "$$DATABASE_URL" up

migrate-down:
	goose -dir ./ledger/migrations postgres "$$DATABASE_URL" down

compose-up:
	docker compose up -d

compose-down:
	docker compose down

logs:
	docker compose logs -f gateway ledger db
