.PHONY: protos

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

protos:
	protoc --proto_path=submodules/protos --go_out=handlers/protos --go-grpc_out=handlers/protos --go_opt=paths=source_relative bulk.proto

bulk:
	docker-compose up -d --no-deps --build bulk
