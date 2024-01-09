# Main API service

Main API service implements the v2 APIs for /codes, /languages, /namespaces, /projects, /snapshots, /hourlys, /docs and /status.

### Getting started:

Need to make sure that `go`, `docker` and `docker-compose` is installed on your machine.

1. Create `.env` file in the project root with following content:

   ```bash
   SERVER_MODE=debug
   AWS_REGION=ap-northeast-1
   AWS_ID=minioadmin
   AWS_KEY=minioadmin
   AWS_URL=http://minio:9000
   AWS_BUCKET=wme-data
   COGNITO_CLIENT_ID=id
   COGNITO_CLIENT_SECRET=secret
   REDIS_ADDR=cache:6379
   REDIS_PASSWORD=
   ACCESS_MODEL=
   ACCESS_POLICY=
   FreeTierGroup=group_1
   ```

1. Start the application by running:

   ```bash
   docker-compose up
   ```

1. Open http://localhost:9202/, login, go to the `buckets` tab and create new bucket called `wme-data`.

1. Copy files from `dataset` directory to the `wme-data` bucket.

### Doing development:

1. To rebuild the service you can run (`main` is a service name, look into `Makefile` for full list of commands):

   ```bash
   make main
   ```

1. If you need to attach log only to one service you can run (`main` is a service name, look into `docker-compose.yaml` for list of services):

   ```bash
   docker-compose logs -f main
   ```

1. Run unit tests:

   ```bash
   go test ./... -v
   ```

1. In order to run linter you need to have `golangci-lint` [installed](https://golangci-lint.run/usage/install/). After that you can run:

   ```bash
   golangci-lint run
   ```
