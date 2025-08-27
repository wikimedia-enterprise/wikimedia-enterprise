.PHONY: up down build test main

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

main:
	docker compose up -d --no-deps --build main
