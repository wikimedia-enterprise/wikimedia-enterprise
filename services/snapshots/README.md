# Wikimedia Enterprise Snapshots

Generates snapshots for Wikimedia Enterprise API(s).

## Getting started

Need to make sure that `go`, `docker` and `docker-compose` are installed on your machine.

1. Create `.env` file in the project root with following content:

   ```bash
   SERVER_PORT=5050
   AWS_URL=http://minio:9000
   AWS_REGION=ap-northeast-1
   AWS_BUCKET=wme-data
   AWS_KEY=password
   AWS_ID=admin
   TOPIC_ARTICLES=local.topic.name
   SCHEMA_REGISTRY_URL=http://schemaregistry:8085
   KAFKA_BOOTSTRAP_SERVERS=broker:29092
   FREE_TIER_GROUP=group_1
   ```

1. Start the application by running:

   ```bash
   docker compose up
   ```

   or

   ```bash
   make up
   ```

1. Create a data bucket using MinIO console:

   1. Go to http://localhost:9000

   1. Login with default credentials:

      | Username | Password   |
      | -------- | ---------- |
      | `admin`  | `password` |

   1. Navigate to `Buckets` > `Create New Bucket`

   1. Enter `wme-data` -or the name you specified for the `AWS_BUCKET` environment variable- as bucket name and click on `Save`

## Developing

1. First you need to go to the [kafka-ui](http://localhost:8380/) and create two topics (this wil make sure you can run `Export` for `enwiki` and `eswiki`):

   1. `aws.structured-data.enwiki-articles-compacted.v1` with 10 partitions

   1. `aws.structured-data.eswiki-articles-compacted.v1` with 10 partitions

1. Then you can generate mock data by doing:

   ```bash
   make mock
   ```

1. To rebuild the service you can run (`snapshots` is a service name, look into `Makefile` for full list of commands):

   ```bash
   make snapshots
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
