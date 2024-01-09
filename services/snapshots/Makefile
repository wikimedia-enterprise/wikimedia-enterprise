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

protos:
	protoc --proto_path=submodules/protos --go_out=handlers/protos --go-grpc_out=handlers/protos --go_opt=paths=source_relative snapshots.proto

mock:
	docker compose up -d --no-deps --build mock
