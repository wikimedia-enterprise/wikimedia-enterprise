# Wikimedia Enterprise Scheduler

Implementation of [Apache AirFlow](https://airflow.apache.org/).

The Scheduler will consume `diffs` and `bulk` services through their gRPC endpoints.

## Getting Started

1. Create `.env` file in the project root with the following content:

   ```shell
   # PostgreSQL configuration
   POSTGRES_USER=airflow
   POSTGRES_PASSWORD=airflow
   POSTGRES_DB=airflow

   # AirFlow configuration
   AIRFLOW__CORE__EXECUTOR=LocalExecutor
   AIRFLOW__CORE__SQL_ALCHEMY_CONN=postgresql+psycopg2://airflow:airflow@postgres/airflow
   AIRFLOW__CORE__FERNET_KEY=''
   AIRFLOW__CORE__DAGS_ARE_PAUSED_AT_CREATION='true'
   AIRFLOW__CORE__LOAD_EXAMPLES='false'
   AIRFLOW__API__AUTH_BACKEND='airflow.api.auth.backend.basic_auth'

   BATCHES_SERVICE_ADDR=host.docker.internal:5050
   BULK_INGESTION_SERVICE_ADDR=host.docker.internal:50051
   SNAPSHOTS_SERVICE_ADDR=host.docker.internal:5050

   # For enabling mutual TLS.
   ROOT_CERTS=
   PRIVATE_KEY=
   CERT_CHAIN=
   ```

   Make sure the PostgreSQL credentials match the credentials in the `AIRFLOW__CORE__SQL_ALCHEMY_CONN` variable.

1. Start the application by running:

   ```bash
   docker-compose up
   # Or if you're using docker-compose >= v2
   docker compose up
   ```

1. Navigate to http://localhost:9000/

1. Log in into AirFlow web UI

   Default credentials:

   | Username  | Password  |
   | --------- | --------- |
   | `airflow` | `airflow` |

   Credentials can be overwritten by adding `AIRFLOW_USERNAME` and `AIRFLOW_PASSWORD` environment variables to the `.env` file.

## Developing

- It is recommended for you to create a virtual environment to manage development dependencies:

  ```bash
  python -m venv .venv
  ```

  This will create a Python virtual environment in the `.venv` directory, that you can activate by running:

  ```bash
  source .venv/bin/activate # for shell, bash, zsh, etc...
  source .venv/bin/activate.fish # for fish
  ```

  To deactivate the Python environment, run:

  ```bash
  deactivate
  ```

- Code style for this project follows [PEP 8](https://www.python.org/dev/peps/pep-0008/) styling guide. To integrate these tools into your development environment run the following, while the virtual environment is activated:

  ```bash
  pip install -r requirements.txt
  ```

  This will install [`black`](https://github.com/psf/black) and [`flake8`](https://flake8.pycqa.org/en/latest/) into your local environment. Then, run:

  ```bash
  black --exclude=protos dags/pipelines/ # for formatting
  flake8 --exclude=protos dags/pipelines/ --max-line-length 160 # for linting
  ```

  VSCode also provides integration with these tools with the [`Python` extension](https://marketplace.visualstudio.com/items?itemName=ms-python.python). To enable them, install the extension, and add the following to your VSCode settings:

  ```js
  "python.formatting.provider": "black",
  "python.linting.flake8Enabled": true,
  ```

---

1. DAG needs some arguments: `projects`, `namespaces`, `exclude_events`. To set these, go to Airflow web UI > click on `Admin` dropdown > click on `Variables`. Click on the `+`. This will take you to `Add Variable` page. Insert an entry for `projects` (Key: projects, Val: copy-paste the contents of file config/projects.json). Hit `Save`. Also, insert entries for `namespaces` (Key: namespaces, Val: [0, 6, 10, 14]), `copy_max_workers`(Key: copy_max_workers, Val: 25), `run_copy`(Key: run_copy, Val: true) and `exclude_events` (Key: exclude_events, Val: ["delete"])

1. To get the Scheduler service running, you can run (`airflow` is a service name, look into `Makefile` for the full list of targets):

   ```bash
   make airflow
   ```

1. If you need to attach log only to one service you can run (`airflow` is a service name, look into `docker-compose.yaml` for list of services):

   ```bash
   docker-compose logs -f airflow
   ```

### Protocol Buffers

1. To build the protobuf files into gRPC code, you will need [`grpcio` and `grpcio-tools`](https://grpc.io/docs/languages/python/quickstart/#prerequisites) installed. Then, run `make protos`. (Check the `Makefile` to see the `protoc` CLI).

   ```bash
   make protos
   ```
