.PHONY: run dev build clean up down swag

run:
	docker compose up

dev:
	make swag
	docker compose -f docker-compose.dev.yml up --build

build:
	docker compose build

swag:
	swag init -g internal/server/api/api.go --output docs

swag-clean:
	rm -rf docs/

clean:
	docker compose down -v
	rm -rf bin/ tmp/

up:
	docker compose up -d

down:
	docker compose down
