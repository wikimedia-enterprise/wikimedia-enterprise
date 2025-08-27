# Wikimedia Enterprise Event Bridge (or Eventstream Listener)

The main purpose of this service is to maintain the connection to event stream.

### Getting started:

Need to make sure that `go`, `docker` and `docker-compose` is installed on your machine.

1. Init `git` sub-modules by running:
    ```bash
    git submodule update --init --remote --recursive
    ```

1. Create `.env` file in the project root with following content:

    ```bash
    KAFKA_BOOTSTRAP_SERVERS=broker:9092
    REDIS_ADDR=redis:6379
    SCHEMA_REGISTRY_URL=http://schemaregistry:8085
    PROMETHEUS_PORT=12411
    OUTPUT_TOPICS='{"version": ["v1"], "service_name": "event-bridge", "location": "aws"}'
    ```

1. Start the application by running:

    ```bash
    docker-compose up
    ```

    After that you should be able to access `redis` UI on [http://localhost:8081/](http://localhost:8081/), and `kafka` UI on [http://localhost:8180/](http://localhost:8180/).


### Doing development:

1. To rebuild the service you can run (`pagechange` is a service name, look into `Makefile` for full list of commands):

    ```bash
    make pagechange
    ```

1. To attach log only to one service you can run:

    ```bash
    sudo docker-compose logs -f pagechange
    ```

1. Run unit tests:

    ```bash
    go test ./... -v
    ```

1. In order to run linter you need to have `golangci-lint` [installed](https://golangci-lint.run/usage/install/). After that you can run:

    ```bash
    golangci-lint run
    ```

## Running the usage example in `example/main.go`:
1. First create `docker-compose.override.yaml` file with following content:
    ```yaml
    version: "3.9"

    services:
      example:
        build:
          context: .
          dockerfile: example/Dockerfile
        depends_on:
          - broker
        env_file:
          - ./.env
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
