# Main API service

Main API service implements the v2 APIs for /codes, /languages, /namespaces, /projects, /snapshots, /hourlys, /docs and /status.

### Getting started:

Need to make sure that `go`, `docker` and `docker-compose` is installed on your machine.

1. Update and init git submodules:

   ```bash
   git submodule update --init --remote --recursive
   ```
2. Create `.env` file in the project root with following content:

   ```bash
   SERVER_MODE=debug
   AWS_REGION=us-east-1
   AWS_ID=minioadmin
   AWS_KEY=minioadmin
   AWS_URL=http://minio:9000
   AWS_BUCKET=wme-primary
   AWS_BUCKET_COMMONS=wme-data
   COGNITO_CLIENT_ID=id
   COGNITO_CLIENT_SECRET=secret
   REDIS_ADDR=cache:6379
   REDIS_PASSWORD=
   ACCESS_MODEL=
   ACCESS_POLICY=
   FreeTierGroup=group_1
   TEST_ONLY_SKIP_AUTH=true
   TEST_ONLY_GROUP=group_3
   SERVER_PORT=4061
   # Enable throttling:
   # CAP_CONFIGURATION='[{"groups":["somegroup"],"limit":5,"products":["articles","structured-contents"],"prefix_group":"cap:ondemand"},{"groups":["newgroup"],"limit":5,"products":["snapshots","structured-snapshots"],"prefix_group":"cap:snapshot"},{"groups":["somegroup"],"limit":15,"products":["chunks"],"prefix_group":"cap:chunk"}]'

   ```
3. Start the application by running:

   ```bash
   docker-compose up
   ```
4. Open http://localhost:9202/, login, go to the `buckets` tab and create new bucket called `wme-primary`.
5. Copy files from `dataset` directory to the `wme-primary` bucket.

### Doing development:

1. To rebuild the service you can run (`main` is a service name, look into `Makefile` for full list of commands):

   ```bash
   make main
   ```
2. If you need to attach log only to one service you can run (`main` is a service name, look into `docker-compose.yaml` for list of services):

   ```bash
   docker-compose logs -f main
   ```
3. To test the main API locally without authentication, comment the following line from `main.go` and add the following line.

   ```golang
     // rtr.Use(httputil.IPAuth(*p.Env.IPAllowList))
     // rtr.Use(httputil.Auth(httputil.NewAuthParams(p.Env.CognitoClientID, p.Cache, p.Provider)))
   	rtr.Use(func(gcx *gin.Context) {
   		gcx.Set("user", &httputil.User{Username: "someuser",
   			Groups: []string{"__replace_this_with_appropriate_group__"}})
   		gcx.Next()
   	})
   ```
4. Run unit tests:

   ```bash
   go test ./... -v
   ```
5. In order to run linter you need to have `golangci-lint` [installed](https://golangci-lint.run/usage/install/). After that you can run:

   ```bash
   golangci-lint run
   ```
