.PHONY: up down build test pagedelete pagemove revisioncreate revisionscore revisionvisibility example

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

pagecreate:
	docker-compose up -d --no-deps --build pagecreate

pagedelete:
	docker-compose up -d --no-deps --build pagedelete

pagemove:
	docker-compose up -d --no-deps --build pagemove

revisioncreate:
	docker-compose up -d --no-deps --build revisioncreate

revisionvisibility:
	docker-compose up -d --no-deps --build revisionvisibility

example:
	docker-compose up -d --no-deps --build example
