.PHONY: up down stop build test realtime mock migrations-init migrations-create migrations-apply migrations-info migrations-validate migrations-destroy-metadata migrations-reset migrate

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

realtime:
	docker compose up -d --no-deps --build realtime

mock:
	docker compose up -d --no-deps --build mock

#############################
# ksqlDB migrations targets #
#############################
ksqldb_container = ksqldb
ksqldb_config = /share/ksql-migrations/ksql-migrations.properties

# Initialize migrations metadata
migrations-init:
	docker compose exec $(ksqldb_container) ksql-migrations -c $(ksqldb_config) initialize-metadata

# Create a migration file with the specified name
migrations-create:
	docker compose exec $(ksqldb_container) ksql-migrations -c $(ksqldb_config) create $(name)

# Apply migrations
migrations-apply:
	docker compose exec $(ksqldb_container) ksql-migrations -c $(ksqldb_config) apply --all

# Get migrations status
migrations-info:
	docker compose exec $(ksqldb_container) ksql-migrations -c $(ksqldb_config) info

# Validate migrations
migrations-validate:
	docker compose exec $(ksqldb_container) ksql-migrations -c $(ksqldb_config) validate

# Destroy migrations metadata. This will not remove already applied migrations
migrations-destroy-metadata:
	docker compose exec $(ksqldb_container) ksql-migrations -c $(ksqldb_config) destroy-metadata

# Reset the migrations state
migrations-reset: migrations-destroy-metadata migrations-init

# Run all needed commands to have an up-to-date version of the migrations
migrate: migrations-reset migrations-apply migrations-validate migrations-info
