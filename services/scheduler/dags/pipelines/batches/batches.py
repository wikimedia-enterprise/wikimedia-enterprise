"""
DAG for Batches generation services. Contains Export and Aggregate tasks
each invoked on an hourly basis.
"""
from datetime import timedelta, datetime, time
from airflow import DAG
from airflow.operators.python import PythonOperator
from logging import getLogger
from grpc import insecure_channel
from pipelines.utils.utils import (
    get_ip_addresses,
    get_channel,
    get_namespaces,
    get_projects,
    get_language,
    get_batch_size,
    chunks,
)

from pipelines.protos.snapshots_pb2_grpc import SnapshotsStub
from pipelines.protos.snapshots_pb2 import (
    ExportRequest,
    AggregateRequest,
)

# Initialize logger
logger = getLogger("airflow.task")

# Initialize service env variable name
addr_name = "BATCHES_SERVICE_ADDR"

# Prefix for s3 storage
prefix = "batches"


def get_start_of_day_timestamp():
    """
    Get start of day timestamp in microseconds,
    note that it subtracts 55 minutes from the current time,
    this is because we want to run the whole previous day at 00:00:00 to avoid data loss.
    """
    return int(
        datetime.combine(datetime.now() - timedelta(minutes=55),
                         time.min).timestamp()
        * 1000
    )


def exports(projects, namespaces, ip_address):
    """
    Calls Batches service Export API for each project-ns
    """

    def export():
        logger.info(
            "Got namespaces configuration: `{namespaces}`".format(
                namespaces=namespaces)
        )
        logger.info(
            "Got projects configuration: `{projects}`\n".format(
                projects=projects)
        )
        logger.info(
            "Trying to get channel on ip address `{ip_address}`\n".format(
                ip_address=ip_address
            )
        )

        since = get_start_of_day_timestamp()

        logger.info("Since parameter value `{since}`\n".format(since=since))

        with insecure_channel(ip_address) as channel:
            for project in projects:
                for namespace in namespaces:
                    logger.info(
                        "Starting an `export` job for project: '{project}' in namespace: '{namespace}'".format(
                            project=project, namespace=namespace
                        )
                    )

                    res = SnapshotsStub(channel).Export(
                        ExportRequest(
                            since=since,
                            prefix=prefix,
                            project=project,
                            namespace=namespace,
                            language=get_language(project),
                        )
                    )

                    logger.info(
                        "Finished an `export` job for project: `{project}` in namespace: `{namespace}` with total: `{total}` and errors: `{errors}`\n".format(
                            project=project,
                            namespace=namespace,
                            total=res.total,
                            errors=res.errors,
                        )
                    )
        return

    return export


def aggregate():
    """
    Call Batches service Aggregate API to generate download links
    for each hourly batch.
    """

    with get_channel(addr_name) as channel:
        logger.info("Starting the `aggregate` job")

        stub = SnapshotsStub(channel)
        res = stub.Aggregate(
            AggregateRequest(since=get_start_of_day_timestamp(), prefix=prefix)
        )

        logger.info(
            "Finished the `aggregate` job with total: `{total}` and errors: `{errors}`\n".format(
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
) as dag:
    task_aggregate = PythonOperator(
        task_id="aggregate",
        python_callable=aggregate,
        execution_timeout=timedelta(days=2),
    )

    ip_addresses = get_ip_addresses(addr_name)
    namespaces = get_namespaces(variable_name="namespaces_batches")
    projects = get_projects(variable_name="projects_batches")
    task_number = 0

    for chunk in chunks(projects, get_batch_size(variable_name="batch_size_batches")):
        task_export = PythonOperator(
            task_id="export_{task_number}".format(task_number=task_number),
            python_callable=exports(
                projects=chunk, namespaces=namespaces, ip_address=next(
                    ip_addresses)
            ),
            execution_timeout=timedelta(hours=3),
        )

        task_aggregate.set_upstream(task_export)
        task_number += 1
