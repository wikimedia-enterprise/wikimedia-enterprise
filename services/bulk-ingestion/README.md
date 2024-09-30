# Wikimedia Enterprise Bulk Ingestion

This service is invoked by a scheduler. There are 3 major tasks performed by this service:

 i) Get all the projects by calling mediawiki actions API. Produce a kafka message for each project.
 ii) Get all the projects by calling mediawiki actions API. Then, get all the namespaces for each project. Produce a kafka message for each namespace.
 iii) For each project and its namespace, get title dumps from mediawiki. Produce a kafka message with all the titles per namespace per project.


### Getting started:

Need to make sure that `go`, `docker` and `docker-compose` is installed on your machine. Also, have protocol buffer compiler and go plugins installed as explained [here](https://grpc.io/docs/languages/go/quickstart/)

1. Init and update `git` sub-modules by running:

    ```bash
    git submodule update --init --remote --recursive
    ```

1. Create `.env` file in the project root with following content:

    ```bash
    KAFKA_BOOTSTRAP_SERVERS=broker:29092
    SERVER_PORT=50051
    MEDIAWIKI_CLIENT_URL=https://en.wikipedia.org/
    SCHEMA_REGISTRY_URL=schemaregistry:8085
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

    After that you should be able to access `kafka` UI on [http://localhost:8280/](http://localhost:8280/).


### Doing development:

1. To regenerate grpc code, run running `make protos`. (Check the `Makefile` to see the `protoc` cli)

1. To rebuild the service, you can run (`bulk` is the service name):

    ```bash
    make bulk
    ```

1. If you need to attach log only to the service, you can run:

    ```bash
    sudo docker-compose logs -f bulk
    ```

1. Run unit tests:

    ```bash
    go test ./... -v
    ```

1. In order to run linter you need to have `golangci-lint` [installed](https://golangci-lint.run/usage/install/). After that you can run:

    ```bash
    golangci-lint run
    ```

1. For dubugging the gPRC server, you may install and use the following gRPC client

   ```bash
   $ go install github.com/fullstorydev/grpcui/cmd/grpcui@latest
   # Start a grpc client to grpc server @ <host>:<port>, with a pointer to a proto file. You can use web UI client then.
   $ grpcui -plaintext -proto submodules/protos/bulk.proto localhost:5050
   gRPC Web UI available at http://127.0.0.1:60551/...
   ```

### Overriding docker configuration:

To override `docker-compose.yaml`, create a file called `docker-compose.override.yaml` and it will be applied automatically.