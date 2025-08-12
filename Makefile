APP := app

.PHONY: run build test tidy lint swag

run:
	docker compose up --build

down:
	docker compose down -v

logs:
	docker compose logs -f app

swag:
	swag init -g cmd/app/main.go -o docs