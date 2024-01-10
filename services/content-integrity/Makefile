.PHONY: up build down stop protos mock ci

####################
# Makefile targets #
####################
up:
	docker compose up -d

build:
	docker compose up -d --build

down:
	docker compose down

stop:
	docker compose stop

mock:
	docker compose up -d --no-deps --build mock

ci:
	docker compose up -d --no-deps --build content-integrity
