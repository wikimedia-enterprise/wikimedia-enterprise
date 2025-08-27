.PHONY: up down build test auth

####################
# Makefile targets #
####################
up:
	docker compose up -d

down:
	docker compose down

build:
	docker compose up -d --build

test:
	go test ./... -v

auth:
	docker compose up -d --no-deps --build auth
