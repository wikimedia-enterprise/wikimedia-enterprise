"""
DAG for Structured-Snapshots services. Contains Export and Aggregate tasks
each invoked on an daily basis.
"""

import sys, time
from os import getenv, path
from datetime import timedelta, datetime
from airflow import DAG
from airflow.models import Variable
from airflow.operators.python import PythonOperator
from airflow.utils.trigger_rule import TriggerRule
from airflow import AirflowException
from logging import getLogger
from grpc import insecure_channel
from airflow.settings import AIRFLOW_HOME
DAGS_HOME=path.join(AIRFLOW_HOME, "dags/git_wme-dags")
WORKERS_HOME=path.join(AIRFLOW_HOME, "dags/dags.git")
sys.path += [DAGS_HOME, WORKERS_HOME]
print(sys.path)
from pipelines.utils.utils import (
    get_projects,
    get_language,
    get_batch_size,
    get_exclude_events,
    chunks,
    slack_failure_callback
)

from pipelines.protos.snapshots_pb2_grpc import SnapshotsStub
from pipelines.protos.snapshots_pb2 import ExportRequest, AggregateRequest

snapshot_address = '0.0.0.0:5051'
environment  = getenv('CLUSTER_ENV')
pod_template_file = path.join(DAGS_HOME, f'pod_templates/structured-snapshots-{environment}.yaml')

# Initialize logger
logger = getLogger("airflow.task")

# Initialize service env variable name
# addr_name = "STRUCTURED_SNAPSHOTS_SERVICE_ADDR"

dag_args = dict(
    dag_id="structured-snapshots",
    start_date=datetime(2022, 1, 1, 0, 0, 0),
    schedule_interval="@daily",
    catchup=False,
    on_failure_callback=slack_failure_callback,
    default_args=dict(
        provide_context=True,
        retries=0,
        retry_delay=timedelta(seconds=30),
    ),
)


def get_since():
    """
    Get since parameter from Airflow Variable.

    Returns:
        int: since parameter
    """
    try:
        return int(Variable.get("structured_snapshots_since"))
    except Exception:
        return -1


def exports(projects, exclude_events, **kwargs):
    """
    Calls Structured-Snapshots service Export API for each project-ns
    """

    logger.info(f"Got projects configuration: `{projects}`")
    logger.info(f"Exclude events of type `{exclude_events}`")

    since = get_since()

    if since > 0:
        logger.info(
            f"Exporting structured-snapshots with since parameter set to `{since}`"
        )
    else:
        logger.info(
            "Exporting structured-snapshots without since parameter, all articles will be exported"
        )

    with insecure_channel(snapshot_address) as channel:
        for project in projects:
            logger.info(f"Starting a `structured export` job for project: '{project}'")

            res = SnapshotsStub(channel).Export(
                ExportRequest(
                    project=project,
                    language=get_language(project),
                    exclude_events=exclude_events,
                    since=since,
                    type="structured",
                    prefix="structured-snapshots",
                    enable_chunking=False
                ),
                wait_for_ready=True
            )

            logger.info(
                f"Finished an `structured export` job for project: `{project}` "
                f" with total: `{res.total}` and errors: `{res.errors}`"
            )

    # Wait for last Prometheus scraping (happens every 15s).
    time.sleep(20)


def aggregate():
    """
    Call Structured-Snapshots service Aggregate API to generate download links
    for each export.
    """
    with insecure_channel(snapshot_address) as channel:
        logger.info("Starting the `structured aggregate` job")

        stub = SnapshotsStub(channel)
        res = stub.Aggregate(
            AggregateRequest(prefix="structured-snapshots"),
            wait_for_ready=True
            )

        logger.info(
            "Finished the `structured aggregate` job with total: `{total}` and errors: `{errors}`\n".format(
                total=res.total,
                errors=res.errors,
            )
        )

with DAG(**dag_args) as dag:
    task_aggregate = PythonOperator(
        task_id="aggregate",
        python_callable=aggregate,
        execution_timeout=timedelta(days=2),
        executor_config = {
            "pod_template_file": pod_template_file
        }
    )

    projects = get_projects(variable_name="structured_projects")
    exclude_events = get_exclude_events()

    for task_number, chunk in enumerate(
        chunks(
            projects, get_batch_size(variable_name="batch_size_structured_snapshots")
        )
    ):
        # opkwargs are used to provide parameters + context on top
        # https://stackoverflow.com/questions/67623479/use-kwargs-and-context-simultaniauly-on-airflow
        task_export = PythonOperator(
            task_id=f"export_{task_number}",
            python_callable=exports,
            op_kwargs=dict(
                projects=chunk,
                namespaces=[0],
                exclude_events=exclude_events
            ),
            execution_timeout=timedelta(days=7),
            executor_config = {
                "pod_template_file": pod_template_file
            }
        )

        # ensure all exports are finished before running the aggregation
        task_aggregate.set_upstream(task_export)
