"""
DAG for Batches generation. Contains Export and Aggregate tasks
each invoked on an hourly basis.
"""
import sys
import os
from time import sleep
from datetime import timedelta, datetime, timezone
from airflow import DAG
from airflow.models import Variable
from airflow.operators.python import PythonOperator
from logging import getLogger
from grpc import insecure_channel
from airflow.settings import AIRFLOW_HOME
from airflow.utils.trigger_rule import TriggerRule

DAGS_HOME = os.path.join(AIRFLOW_HOME, "dags/git_wme-dags")
WORKERS_HOME = os.path.join(AIRFLOW_HOME, "dags/dags.git")
sys.path += [DAGS_HOME, WORKERS_HOME]

from pipelines.utils.utils import (
    get_namespaces,
    get_projects,
    get_language,
    get_batch_size,
    chunks,
    retry_failed_tasks_callback,
    run_retryable,
    send_msg_to_slack,
    MAX_CUSTOM_RETRIES
)

from pipelines.protos.snapshots_pb2_grpc import SnapshotsStub
from pipelines.protos.snapshots_pb2 import (
    ExportRequest,
    AggregateRequest,
)

# Initialize logger
logger = getLogger("airflow.task")

# Initialize service env variable name
service_address = "0.0.0.0:5051"
environment = os.getenv('CLUSTER_ENV')
pod_template_file = os.path.join(DAGS_HOME, f'pod_templates/batches-{environment}.yaml')

# Prefix for s3 storage
prefix = "batches"


def start_of_last_hour_utc():
    now = datetime.now(timezone.utc)
    # Round down to the start of the current hour.
    current_hour_start = now.replace(minute=0, second=0, microsecond=0)
    last_hour_start = current_hour_start - timedelta(hours=1)
    return last_hour_start


def exports(projects, namespaces):
    """
    Calls Batches service Export API for each project-ns
    """

    def export(**context):
        # Will skip failing projects/namespaces in `export` until `max_export_errors` errors are observed,
        # at which point the `export` task will finish to unblock `aggregate`.
        max_export_errors = Variable.get("max_export_errors_per_task", 20)
        errors = 0

        logger.info(
            "Got namespaces configuration: `{namespaces}`".format(namespaces=namespaces)
        )
        logger.info(
            "Got projects configuration: `{projects}`\n".format(projects=projects)
        )

        since = int(start_of_last_hour_utc().timestamp() * 1000)
        logger.info("Since parameter value `{since}`\n".format(since=since))

        can_skip_errors = False
        ti = context.get("ti")
        if ti is None:
            custom_try = 1
            logger.warning("Could not obtain task instance")
        else:
            custom_try = ti.try_number

        # can_skip_errors is True on the last attempt, the attempt we will *not* retry even if it fails.
        # In that attempt, we are going to try to let aggregate run to collect whatever succeeded from the
        # exports.
        can_skip_errors = custom_try > MAX_CUSTOM_RETRIES

        processed = 0
        with insecure_channel(service_address) as channel:
            stub = SnapshotsStub(channel)
            for project in projects:
                if errors >= max_export_errors:
                    break

                for namespace in namespaces:
                    logger.info(
                        "Starting an `export` job for project: '{project}' in namespace: '{namespace}'".format(
                            project=project, namespace=namespace
                        )
                    )

                    def export_project_namespace():
                        res = stub.Export(
                            ExportRequest(
                                since=since,
                                prefix=prefix,
                                project=project,
                                namespace=namespace,
                                language=get_language(project),
                                enable_chunking=False,
                                enable_non_cumulative_batches=True,
                            ),
                            wait_for_ready=True
                        )

                        logger.info(
                            "Finished an `export` job for project: `{project}` in namespace: `{namespace}` with total: `{total}` and errors: `{errors}`\n".format(
                                project=project,
                                namespace=namespace,
                                total=res.total,
                                errors=res.errors,
                            )
                        )

                    # Assume all errors are retryable for now and refine as we observe them.
                    try:
                        run_retryable(export_project_namespace, lambda _: True)

                        processed += 1
                    except Exception as e:
                        msg = f"Error found during `export` for project `{project}` in namespace `{namespace}`: {e}"
                        if not can_skip_errors:
                            raise Exception(msg)
                        logger.error(msg)
                        errors += 1

                        if errors >= max_export_errors:
                            break

        # Wait for last Prometheus scraping (happens every 15s).
        sleep(20)

        if errors > 0:
            num_combinations = len(projects) * len(namespaces)
            skipped = num_combinations - processed
            msg = f"Task succeeded with errors. Processed: {processed}. Skipped: {skipped}."
            logger.error(msg)
            send_msg_to_slack(context, msg)

    return export


def aggregate():
    """
    Call Batches service Aggregate API to generate download links
    for each hourly batch.
    """

    with insecure_channel(service_address) as channel:
        logger.info("Starting the `aggregate` job")

        ts = start_of_last_hour_utc()
        since = int(ts.timestamp() * 1000)
        aggregate_prefix = "%s/%s/%s" % (prefix, ts.strftime("%Y-%m-%d"), ts.strftime('%H'))

        stub = SnapshotsStub(channel)
        res = stub.Aggregate(
            AggregateRequest(since=since, prefix=aggregate_prefix),
            wait_for_ready=True
        )

        logger.info(
            "Finished the `aggregate` task for `{aggregate_prefix}` job with total: `{total}` and errors: `{errors}`\n".format(
                aggregate_prefix=aggregate_prefix,
                total=res.total,
                errors=res.errors,
            )
        )
    return


with DAG(
    dag_id="batches",
    start_date=datetime(2022, 1, 1, 0, 0, 0),
    schedule_interval="@hourly",
    catchup=False,
    dagrun_timeout=timedelta(minutes=55),
    on_failure_callback=retry_failed_tasks_callback,
    default_args=dict(
        provide_context=True,
        retries=0,
        retry_delay=timedelta(seconds=30),
    ),
) as dag:
    task_aggregate = PythonOperator(
        task_id="aggregate",
        python_callable=aggregate,
        execution_timeout=timedelta(minutes=40),
        executor_config={
            "pod_template_file": pod_template_file
        },
    )
    namespaces = get_namespaces(variable_name="namespaces_batches")
    projects = get_projects(variable_name="projects_batches")
    task_number = 0

    for chunk in chunks(projects, get_batch_size(variable_name="batch_size_batches")):
        task_export = PythonOperator(
            task_id="export_{task_number}".format(task_number=task_number),
            python_callable=exports(
                projects=chunk, namespaces=namespaces
            ),
            execution_timeout=timedelta(minutes=40),
            executor_config={
                "pod_template_file": pod_template_file
            },
        )

        task_aggregate.set_upstream(task_export)
        task_number += 1
