# Snapshots Service

Description goes here...

## Getting started

Need to make sure that `go`, `docker` and `docker-compose` are installed on your machine.

1. Init `git` sub-modules by running:

   ```bash
   git submodule update --init --remote --recursive
   ```

1. Create `.env` file in the project root with following content:

   ```bash
   SERVER_PORT=5050
   AWS_URL=http://minio:9000
   AWS_REGION=ap-northeast-1
   AWS_BUCKET=wme-primary
   AWS_BUCKET_COMMONS=wme-data
   AWS_KEY=password
   AWS_ID=admin
   TOPIC_ARTICLES=local.topic.name
   SCHEMA_REGISTRY_URL=http://schemaregistry:8085
   KAFKA_BOOTSTRAP_SERVERS=broker:29092
   FREE_TIER_GROUP=group_1
   # For structured-contents snapshots:
   # TOPICS={"service_name": "structured-contents"}
   ```

1. Start the application by running:

   ```bash
   docker compose up
   ```

   or

   ```bash
   make up
   ```

1. Create a primary bucket using MinIO console:

   1. Go to http://localhost:9000

   1. Login with default credentials:

      | Username | Password   |
      | -------- | ---------- |
      | `admin`  | `password` |

   1. Navigate to `Buckets` > `Create New Bucket`

   1. Enter `wme-primary` -or the name you specified for the `AWS_BUCKET` environment variable- as bucket name and click on `Save`. If you run the `mock` image, it will create this bucket automatically.

   1. Enter `wme-data` -or the name you specified for the `AWS_BUCKET_COMMONS` environment variable- as bucket name and click on `Save`

## Developing

1. You can set up the Kafka topics and generate mock data by running:

   ```bash
   make mock
   ```

   You can make changes to them in the [kafka-ui](http://localhost:8380/).

1. To rebuild the service you can run (`snapshots` is a service name, look into `Makefile` for full list of commands):

   ```bash
   make snapshots # or make snapshotsX on Apple Silicon
   ```

1. If you need to attach log only to one service you can run (`exports` is a service name, look into `docker-compose.yaml` for list of services):

   ```bash
   docker-compose logs -f snapshots
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
   $ grpcui -plaintext -proto submodules/protos/snapshots.proto localhost:5050
   gRPC Web UI available at http://127.0.0.1:60551/...
   ```

   Alternatively, you can send requests from the CLI using [`grpcurl`](https://github.com/fullstorydev/grpcurl):

   ```bash
   $ grpcurl -plaintext -d @ -proto submodules/protos/snapshots.proto localhost:5050 snapshots.Snapshots.Export < mock/export-grpc.json
   $ grpcurl -plaintext -d @ -proto submodules/protos/snapshots.proto localhost:5050 snapshots.Snapshots.Export < mock/export-sc-grpc.json

   # Copy request
   $ grpcurl -plaintext -d @ -proto submodules/protos/snapshots.proto localhost:5050 snapshots.Snapshots.Copy < mock/copy-grpc.json

   # AggregateCopy request
   $ grpcurl -plaintext -d @ -proto submodules/protos/snapshots.proto localhost:5050 snapshots.Snapshots.AggregateCopy < mock/aggregatecopy-grpc.json
   ```
   To test group_1 chunking behaviour in dev, include `run_copy = true` env var in dags

