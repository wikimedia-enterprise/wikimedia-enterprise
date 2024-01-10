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

bulk:
	docker-compose up -d --no-deps --build bulk
