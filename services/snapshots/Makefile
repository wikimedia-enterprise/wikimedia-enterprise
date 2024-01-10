.PHONY: snapshots up build down stop protos mock

####################
# Makefile targets #
####################
snapshots:
	docker compose up -d --no-deps --build snapshots

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
