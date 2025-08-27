# Wikimedia Enterprise DAGs

The DAGS consume `snapshots`, `bulk-ingestion` and other services through gRPC endpoints.

## Getting Started EKS Workflow
Unlike Airflow ECS, our K8s infrastructure doesn't use `dev` and `main` branches. Instead, we use different Helm values for `dv` and `pr` environments that can point to different Git branches.

The steps below outline how to test DAGs using the `dv` K8s cluster. Note that local testing isn't possible, as our Python DAG code use a K8s provider to spin up Airflow worker instances.

## Steps to QA test you DAG updates

1. Create a feature branch in the `dags` repo and update the DAG code.
2. Push your feature branch to the remote `dags` repo.
3. Create a feature branch in `argocd-apps` repo, to configure the `dv` K8s cluster to use your feature branch DAGs.
4. Edit `values-wme-eks-dv.yaml`, changing `branch: main` to point to your `dags` feature branch, example: `branch: feature-T12345-My-cool-stuff`. [Helm code](image.png)
   - WARNING: Don't update the `values-wme-eks-pr.yaml`  unless you want the changes in production!!
5. Also, modify the `pod-templates` YAML for the DAG you are updating, so ArgoCD knows which branch to sync. For example if you are QA testing snapshot DAG changes in `dv`, then edit `pod_templates/snapshot-dv.yaml` and change the value under property `GITSYNC_BRANCH` to the Git branch you are working on the `dags` repo.  Example:

```YAML
Change from Main to a feature branch:
From:
        - name: GITSYNC_BRANCH
          value: "main"
To:
        - name: GITSYNC_BRANCH
          value: "feature/T389546-add-group_1-snapshot-chunking-refresh"
```
6. Push your `argocd-apps` feature branch, create the MR, and request team review the Helm configuration change.
7. After approval and merge.
   - Verify that ArgoCD deployment updated correctly.
   - Verify that Airflow `dv` has your latest DAG code changes.
8. Then QA test your DAGs in [Airflow UI](https://kafka-ui-dev.wikimediaenterprise.org/) in the `dv` K8s cluster.
9. When QA tests are successful, create a MR to merge your `dags` changes.
10. Revert the Helm changes in `values-wme-eks-dv.yaml` and the pod-templates YAML file, so they points back to the Git `main` branch.
   - After changes both `dv` and `pr` should point to the `main` `dags` branch.

11. Once the `dags` changes are merged to `main`, your DAG code changes will appear in production [Airflow UI](https://airflow-prod.wikimediaenterprise.org/).

## Getting Started

1. Init and update `git` submodules by running:

   ```bash
   git submodule update --init --remote --recursive
   ```

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
  flake8 --exclude dags/pipelines/protos/ --max-line-length 160 --extend-ignore=W293,W291,W391,E203 # for linting
  ```

  VSCode also provides integration with these tools with the [`Python` extension](https://marketplace.visualstudio.com/items?itemName=ms-python.python). To enable them, install the extension, and add the following to your VSCode settings:

  ```js
  "python.formatting.provider": "black",
  "python.linting.flake8Enabled": true,
  ```

### Protocol Buffers

1. To build the protobuf files into gRPC code, you will need [`grpcio` and `grpcio-tools`](https://grpc.io/docs/languages/python/quickstart/#prerequisites) installed. Then, run `make protos`. (Check the `Makefile` to see the `protoc` CLI).

   ```bash
   make protos
   ```

