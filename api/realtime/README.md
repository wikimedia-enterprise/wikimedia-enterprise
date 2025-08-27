# Realtime API

This service exposes API endpoints that allow clients to subscribe to article topics from the structured data service. For each API call, a kafka consumer is created and article event messages are streamed to this consumer.

### Getting started:

Need to make sure that `go`, `docker` and `docker-compose` is installed on your machine.

1. Init and update `git` sub-modules by running:

   ```bash
   git submodule update --init --remote --recursive
   ```

1. Create `.env` file in the project root with following content:

   ```bash
   KSQL_URL=http://ksqldb:8088
   KAFKA_BOOTSTRAP_SERVERS=broker:29092
   SCHEMA_REGISTRY_URL=http://schemaregistry:8085
   SERVER_MODE=release
   SERVER_PORT=4040
   AWS_REGION=us-east-1
   AWS_ID=some_id
   AWS_KEY=some_secret
   COGNITO_CLIENT_ID=your_cognito_id
   COGNITO_CLIENT_SECRET=
   REDIS_ADDR=cache:6379
   REDIS_PASSWORD=
   LOG_LEVEL=debug
   ACCESS_MODEL="[request_definition]
   r = sub, obj, act

   [policy_definition]
   p = sub, obj, act

   [role_definition]
   g = _, _

   [policy_effect]
   e = some(where (p.eft == allow))

   [matchers]
   m = (g(r.sub, p.sub) || keyMatch(r.sub, p.sub)) && keyMatch2(r.obj, p.obj) && regexMatch(r.act, p.act)
   "
   ACCESS_POLICY="p, realtime, /v2/articles, GET
   p, realtime, /v2/articles, POST
   g, group_3, realtime
   "
   ```

1. Start the application by running:

   ```bash
   docker compose up
   ```

1. After that make sure to run migrations by running:

   ```bash
   make migrate
   ```

### Doing development:

1. To rebuild the service you can run (`realtime` is a service name, look into `Makefile` for full list of commands):

   ```bash
   make realtime
   ```

1. If you need to attach log only to one service you can run (`realtime` is a service name, look into `docker-compose.yaml` for list of services):

   ```bash
   sudo docker compose logs -f realtime
   ```

1. Run unit tests:

   ```bash
   go test ./... -v
   ```

1. In order to run linter you need to have `golangci-lint` [installed](https://golangci-lint.run/usage/install/). After that you can run:

   ```bash
   golangci-lint run
   ```

### Running the producer mock service:

1. To generate mock data from `mock` directory (you can find `*.ndjson` files in the directory):

   ```bash
   docker compose up --build mock
   ```

   If you need to produce mock events in several partitions, first update the topic partitions from `kafka-ui` under topic settings. Then run the docker command as follows with partitions specified.

   ```bash
   docker-compose run mock ./main -p=0,1,2,3
   ```

1. To view logs you can run:

   ```bash
   docker compose logs -f mock
   ```

### Adding ksqlDB CLi client:

1. First create `docker-compose.override.yaml` file with following content:

   ```yaml
   version: "3.9"

   services:
     ksqldb-cli:
       image: confluentinc/ksqldb-cli:latest
       container_name: ksqldb-cli
       depends_on:
         - ksqldb
       entrypoint: /bin/sh
       tty: true
   ```

1. Now you can access the cli by running:

   ```bash
   docker compose exec ksqldb-cli ksql http://ksqldb:8088
   ```

### Updating the stream schema

In the scenario when you need to roll out new schema version follow the next steps:

1. Create new migration using this command that will create a new version of the stream (don't forget in `ksqldb/ksql-migrations.properties` temporary change `ksql.server.url` to `http://0.0.0.0:8088`, revert it back after new migration was created):

   ```bash
   make migrations-create name=[stream]_[version]
   ```

1. Update the handlers to use your newly created stream. Need to update the default value of `ARTICLES_STREAM` env variable.

1. If there are more than two versions of the stream, please also drop the oldest version of the stream so that we have only two versions of the stream available at any point in time.
