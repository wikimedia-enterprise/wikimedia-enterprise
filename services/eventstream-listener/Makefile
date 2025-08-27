.PHONY: up down stop build test pagechange

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

pagechange:
	docker compose up -d --no-deps --build pagechange
