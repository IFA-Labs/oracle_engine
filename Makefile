.PHONY: run dev build clean up down

run:
	docker-compose up

dev:
	docker-compose -f docker-compose.dev.yml up --build

build:
	docker-compose build

clean:
	docker-compose down -v
	rm -rf bin/ tmp/

up:
	docker-compose up -d

down:
	docker-compose down
