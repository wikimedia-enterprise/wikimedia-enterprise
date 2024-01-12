# Wikimedia Enterprise On-Demand

Stores each article and version in a structured way, reflects changes from different kafka topics.

### Getting started:

Need to make sure that `go`, `docker` and `docker-compose` is installed on your machine.

1. Create `.env` file in the project root with following content:

   ```bash
   KAFKA_BOOTSTRAP_SERVERS=broker:29092
   KAFKA_CONSUMER_GROUP_ID=ondemand
   SCHEMA_REGISTRY_URL=http://schemaregistry:8085
   AWS_URL=http://minio:9000
   AWS_REGION=ap-northeast-1
   AWS_BUCKET=wme-data
   AWS_KEY=password
   AWS_ID=admin
   ```

1. Start the application by running:

   ```bash
   docker-compose up
   ```

   After that you should be able to access `kafka` UI on [http://localhost:8180/](http://localhost:8180/).

### Doing development:

1. To rebuild the service you can run (`example` is a service name, look into `Makefile` for full list of commands):

   ```bash
   make example
   ```

1. If you need to attach log only to one service you can run (`example` is a service name, look into `docker-compose.yaml` for list of services):

   ```bash
   sudo docker-compose logs -f example
   ```

1. Run unit tests:

   ```bash
   go test ./... -v
   ```

1. In order to run linter you need to have `golangci-lint` [installed](https://golangci-lint.run/usage/install/). After that you can run:

   ```bash
   golangci-lint run
   ```

1. To produce mock messages use (to update mock data use `.ndjson` files in the `/mock` directory):

   ```bash
   make mock
   ```

### Overriding docker configuration:

To override `docker-compose.yaml` just create a file called `docker-compose.override.yaml` and it will be applied automatically.
