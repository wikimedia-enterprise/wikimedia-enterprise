# Wikimedia Enterprise Authentication API

This service exposes API endpoints for managing login and tokens. In the current implementation, for each API call, a call is made to AWS cognito service that performs the actual authentication management.

### Getting started:

Need to make sure that `go`, `docker` and `docker-compose` is installed on your machine.

1. Create `.env` file in the project root with following content:

   ```bash
   AWS_REGION=us-east-1
   AWS_ID=some_id
   AWS_KEY=some_secret
   COGNITO_CLIENT_ID=your_cognito_id
   COGNITO_SECRET=your_secret
   SERVER_MODE=release
   SERVER_PORT=4050
   COGNITO_USER_POOL_ID=user-pool
   COGNITO_USER_GROUP=free-tier
   REDIS_ADDR=redis:6379
   REDIS_PASSWORD=password
   ACCESS_POLICY=policy-csv
   GROUP_DOWNLOAD_LIMIT=10000
   ```

1. Start the application by running:

   ```bash
   docker compose up
   ```

### Doing development:

1. To rebuild the service you can run (`auth` is a service name, look into `Makefile` for full list of commands):

   ```bash
   make auth
   ```

1. If you need to attach log only to one service you can run (`auth` is a service name, look into `docker-compose.yaml` for list of services):

   ```bash
   sudo docker compose logs -f auth
   ```

1. Run unit tests:

   ```bash
   go test ./... -v
   ```

1. In order to run linter you need to have `golangci-lint` [installed](https://golangci-lint.run/usage/install/). After that you can run:

   ```bash
   golangci-lint run
   ```
