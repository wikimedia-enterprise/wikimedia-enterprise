.PHONY: up down stop build test structured mock

####################
# Makefile targets #
####################
up:
	docker compose up -d

down:
	docker compose down

stop:
	docker compose stop

build:
	docker compose up -d --build

test:
	go test ./... -v

structured:
	docker compose up -d --no-deps --build structured

mock:
	docker compose up -d --no-deps --build mock
