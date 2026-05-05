# Wikimedia Enterprise Structured Contents

This service works on the kafka topic messages from structured data service. Based on the type of event message, the service handler calls the parser APIs and wmf API. The purpose of these calls are to get parsed contents such as infoboxes, article sections, etc. Finally, the structured contents service handler produce their own kafka events to `aws.structured-contents.{project}-articles-compacted.v1` compacted topic and to a `aws.structured-contents.articles.v1` (_update topic_; time-based deletion configured topic). 

### Getting started:

Need to make sure that `go`, `docker` and `docker-compose` is installed on your machine.

1. Init `git` sub-modules by running:

   ```bash
   git submodule update --init --remote --recursive
   ```

1. Create `.env` file in the project root with following content:

   ```bash
   KAFKA_BOOTSTRAP_SERVERS=broker:29092
   KAFKA_CONSUMER_GROUP_ID=unique
   SCHEMA_REGISTRY_URL=http://schemaregistry:8085
   ```

1. Concurrency configuration, if you need to adjust the default concurrency values you can use `NUMBER_OF_WORKERS` variable (default number of workers is `5`).

1. Start the application by running:

   ```bash
   docker compose up
   ```

   After that you should be able to access `kafka` UI on [http://localhost:8180/](http://localhost:8180/).

### Doing development:

1. To rebuild the service you can run:

   ```bash
   make structured
   ```

1. If you need to attach log to a service you can run:

   ```bash
   sudo docker-compose logs -f structured
   ```

1. Run unit tests:

   ```bash
   go test ./... -v
   ```

1. In order to run linter you need to have `golangci-lint` [installed](https://golangci-lint.run/usage/install/). After that you can run:

   ```bash
   golangci-lint run
   ```

### Mocking topic input:

1. Then you need to run:

   ```bash
   docker compose up --build mock
   ```

### Overriding docker configuration:

To override `docker-compose.yaml` just create a file called `docker-compose.override.yaml` and it will be applied automatically.
