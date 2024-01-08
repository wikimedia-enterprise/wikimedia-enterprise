.PHONY: up down build test articlebulk articledelete articleupdate articlevisibility mock example protos

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

articlebulk:
	docker compose up -d --no-deps --build articlebulk

articledelete:
	docker compose up -d --no-deps --build articledelete

articledeleterror:
	docker compose up -d --no-deps --build articledeleterror

articleupdate:
	docker compose up -d --no-deps --build articleupdate

articleupdaterror:
	docker compose up -d --no-deps --build articleupdaterror

articlevisibility:
	docker compose up -d --no-deps --build articlevisibility

mock:
	docker compose up -d --no-deps --build mock

example:
	docker compose up -d --no-deps --build example

protos:
	protoc --proto_path=submodules/protos --go_out=packages/textprocessor --go-grpc_out=packages/textprocessor --go_opt=paths=source_relative textprocessor.proto
	protoc --proto_path=submodules/protos --go_out=packages/contentintegrity --go-grpc_out=packages/contentintegrity --go_opt=paths=source_relative contentintegrity.proto
