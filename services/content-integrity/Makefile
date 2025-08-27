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

protos:
	protoc --proto_path=submodules/protos --go_out=handlers/server/protos --go-grpc_out=handlers/server/protos --go_opt=paths=source_relative contentintegrity.proto

mock:
	docker compose up -d --no-deps --build mock

ci:
	docker compose up -d --no-deps --build content-integrity
