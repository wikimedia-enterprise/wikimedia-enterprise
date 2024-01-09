# Wikimedia Enterprise Content Integrity Service

Service that exposes Credibility Signals to our internal infrastructure through gRPC calls.

### Getting started:

Need to make sure that `go`, `docker` and `docker-compose` is installed on your machine.

1. Create `.env` file in the project root with following content:

   ```bash
   KAFKA_BOOTSTRAP_SERVERS=broker:29092
   PROMETHEUS_PORT=12411
   SERVER_PORT=5051
   REDIS_ADDR=redis:6379
   SCHEMA_REGISTRY_URL=http://schemaregistry:8085
   TOPIC_ARTICLE_CREATE="aws.event-bridge.article-create.v1"
   TOPIC_ARTICLE_UPDATE="aws.event-bridge.article-update.v1"
   TOPIC_ARTICLE_MOVE="aws.event-bridge.article-move.v1"
   BREAKING_NEWS_KEYS_EXPIRATION=48
   ```

1. Start the application by running:

   ```bash
   docker-compose up
   ```

### Mocking topic input:

1. Then you need to run:

   ```bash
   docker-compose up --build mock
   ```

1. To view logs you can run:

   ```bash
   docker-compose logs -f mock
   ```

1. Then you can run to start the container by running:

   ```bash
   docker compose up --build
   ```

1. To view logs you can run:
   ```bash
   docker compose logs -f example
   ```

### Overriding docker configuration:

To override `docker-compose.yaml` just create a file called `docker-compose.override.yaml` and it will be applied automatically.

### Developing

1. Then you can generate mock data by doing:

   ```bash
   make mock
   ```

1. To rebuild the service you can run (`content-integrity` is a service name, look into `Makefile` for full list of commands):

   ```bash
   make ci
   ```

1. If you need to attach log only to one service you can run (`content-integrity` is a service name, look into `docker-compose.yaml` for list of services):

   ```bash
   docker compose logs -f content-integrity
   ```

1. Run unit tests:

   ```bash
   go test ./... -v
   ```

1. In order to run linter you need to have `golangci-lint` [installed](https://golangci-lint.run/usage/install/). After that you can run:

   ```bash
   golangci-lint run
   ```
