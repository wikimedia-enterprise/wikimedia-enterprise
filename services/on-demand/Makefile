.PHONY: up down stop build test on-demand mock

####################
# Makefile targets #
####################
up:
	docker-compose up -d

down:
	docker-compose down

stop:
	docker-compose stop

build:
	docker-compose up -d --build

test:
	go test ./... -v

example:
	docker-compose up -d --no-deps --build example

mock:
	sudo docker compose up -d --build mock
