# Wikimedia Enterprise Structured Data

This service works on the kafka topic messages from event bridge service. Based on the type of event message, the service handlers make calls to wikimedia's REST Base, Actions and ORES API. The purpose of these calls are to get html, metadata, validate the accuracy of event action, etc. Finally, the structured data service handlers produce their own kafka events.
If wikimedia API call fails, an event will be published to an error topic, which is used in <>error service to reprocess. In case of multiple failure of a message, it will be published to a dead letter topic.

### Getting started:

Need to make sure that `go`, `docker` and `docker-compose` is installed on your machine.

1. Create `.env` file in the project root with following content:

   ```bash
   KAFKA_BOOTSTRAP_SERVERS=broker:29092
   KAFKA_CONSUMER_GROUP_ID=unique
   SCHEMA_REGISTRY_URL=http://schemaregistry:8085
   CONTENT_INTEGRITY_URL=content-integrity:5051
   BREAKING_NEWS_ENABLED=true
   TOPICS={"articles_compacted":"aws.structured-data.articles-compacted.v1","files_compacted":"aws.structured-data.files-compacted.v1","templates_compacted":"aws.structured-data.templates-compacted.v1","categories_compacted":"aws.structured-data.categories-compacted.v1"}
   PARTITION_CONFIG={"articles_compacted":"aws.structured-data.articles-compacted.v1","files_compacted":"aws.structured-data.files-compacted.v1","templates_compacted":"aws.structured-data.templates-compacted.v1","categories_compacted":"aws.structured-data.categories-compacted.v1"}
   ```

   Note: Since the same handler is being used in article-X and article-X-error service, the handler needs to know where to consume from. For instance, `articledelete` service consumes from aws.event-bridge.article-delete.v1. And, articledeleterror service would consume from aws.structured-data.article-delete-error.v1. This is controlled by using `TOPIC_ARTICLE_DELETE` env variable. The default value is aws.event-bridge.article-delete.v1. We override this env variable to aws.structured-data.article-delete-error.v1 in `articledeleterror` service docker config.

1. Concurrency configuration, if you need to adjust the default concurrency values you can use `NUMBER_OF_WORKERS` variable (default number of workers is `5`).

1. Start the application by running:

   ```bash
   docker-compose up
   ```

   After that you should be able to access `kafka` UI on [http://localhost:8180/](http://localhost:8180/).

### Doing development:

1. To rebuild the service you can run (`articledelete` is a service name, look into `Makefile` for full list of commands):

   ```bash
   make articledelete
   ```

1. If you need to attach log only to one service you can run (`articledelete` is a service name, look into `docker-compose.yaml` for list of services):

   ```bash
   sudo docker-compose logs -f articledelete
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
