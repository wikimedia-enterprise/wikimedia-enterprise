"""
DAG for Snapshots services. Contains Export and Aggregate tasks
each invoked on an daily basis.
"""
import sys
import time
from os import getenv, path
from datetime import timedelta, datetime
from airflow import DAG
from airflow.models import Variable
from airflow.operators.python import PythonOperator
from airflow import AirflowException

from logging import getLogger
from grpc import insecure_channel

from airflow.settings import AIRFLOW_HOME
DAGS_HOME = path.join(AIRFLOW_HOME, "dags/git_wme-dags")
WORKERS_HOME = path.join(AIRFLOW_HOME, "dags/dags.git")
sys.path += [DAGS_HOME, WORKERS_HOME]
print(sys.path)

from pipelines.utils.utils import (
    get_namespaces,
    get_projects,
    get_language,
    get_total_articles,
    get_copy_workers,
    run_copy,
    get_exclude_events,
    get_chunks_by_project_size,
    ecs_service_set_task_count,
    verify_ip_addresses_returned_by_dns,
    get_desired_count,
    slack_failure_callback,
    retry_failed_tasks_callback
)

from pipelines.protos.snapshots_pb2_grpc import SnapshotsStub
from pipelines.protos.snapshots_pb2 import (
    ExportRequest,
    CopyRequest,
    AggregateRequest,
    AggregateCopyRequest,
)

snapshot_address = '0.0.0.0:5051'
environment = getenv('CLUSTER_ENV')
pod_template_file = path.join(DAGS_HOME, f'pod_templates/snapshot-{environment}.yaml')

# Initialize logger
logger = getLogger("airflow.task")

# Initialize service env variable name
addr_name = "SNAPSHOTS_SERVICE_ADDR"

default_args = {"provide_context": True}
enable_retries = False
if enable_retries:
    default_args["retries"] = 1
    default_args["retry_delay"] = timedelta(seconds=30)

dag_args = dict(
    dag_id="snapshots",
    start_date=datetime(2022, 1, 1, 0, 0, 0),
    schedule_interval="@daily",
    catchup=False,
    on_failure_callback=retry_failed_tasks_callback,
    default_args=default_args,
)

chunking_enabled = int(Variable.get("chunking_enabled", default_var=True))


def get_since():
    """
    Get since parameter from Airflow Variable.

    Returns:
        int: since parameter
    """
    try:
        return int(Variable.get("snapshots_since"))
    except Exception:
        return -1


def exports(
    projects, namespaces, exclude_events, **kwargs
):
    """
    Calls Snapshots service Export API for each project-ns
    """
    logger.info(f"Got namespaces configuration: `{namespaces}`")
    logger.info(f"Got projects configuration: `{projects}`")
    logger.info(f"Exclude events of type `{exclude_events}`")
    logger.info(f"Trying to get channel on ip address `{snapshot_address}`")

    since = get_since()

    if since > 0:
        logger.info(f"Exporting snapshots with since parameter set to `{since}`")
    else:
        logger.info(
            "Exporting snapshots without since parameter, all articles will be exported"
        )

    with insecure_channel(snapshot_address) as channel:
        for project in projects:
            logger.info(f'These are the projects: {project}')
            for namespace in namespaces:
                logger.info(
                    f"Starting an `export` job for project: '{project}' in namespace: '{namespace}'"
                )

                res = SnapshotsStub(channel).Export(
                    ExportRequest(
                        project=project,
                        namespace=namespace,
                        language=get_language(project),
                        exclude_events=exclude_events,
                        since=since,
                        type="article",
                        enable_chunking=chunking_enabled,
                    ),
                    wait_for_ready=True
                )

                logger.info(
                    f"Finished an `export` job for project: `{project}` in namespace: "
                    f"`{namespace}` with total: `{res.total}` and errors: `{res.errors}`"
                )

    # Wait for last Prometheus scraping (happens every 15s).
    time.sleep(20)


def copy(
    workers, projects, namespaces, datetime, **kwargs
):
    """
    Calls Snapshots service Copy API for a ns, and a set of projects.
    """
    if datetime.day in [2, 21] or run_copy():
        logger.info(f"Got namespaces configuration: `{namespaces}`")
        logger.info(f"Got projects configuration: `{projects}`")
        logger.info(f"Got workers configuration: `{workers}`")
        logger.info(f"Trying to get channel on ip address `{snapshot_address}`")

        with insecure_channel(snapshot_address) as channel:
            for namespace in namespaces:
                logger.info(
                    f"Starting a `copy` job for namespace: '{namespace}' for projects '{projects}'"
                )

                res = SnapshotsStub(channel).Copy(
                    CopyRequest(workers=workers, projects=projects, namespace=namespace),
                    wait_for_ready=True
                )

                logger.info(
                    f"Finished a `copy` job for projects: `{projects}` in namespace: "
                    f"`{namespace}` with total: `{res.total}` and errors: `{res.errors}`"
                )
    else:
        logger.info("Not the first of the month. Skipping copy task.")


def aggregate():
    """
    Call Snapshots service Aggregate API to generate download links
    for each export.
    """
    with insecure_channel(snapshot_address) as channel:
        logger.info("Starting the `aggregate` job")

        stub = SnapshotsStub(channel)
        res = stub.Aggregate(
            AggregateRequest(),
            wait_for_ready=True
        )

        logger.info(
            f"Finished the `aggregate` job with total: `{res.total}` and errors: `{res.errors}`"
        )


def aggregate_copy(projects, namespaces, datetime):
    """
    Call Snapshots service AggregateCopy API to copy aggregated metadata of snapshots.
    """

    if not (datetime.day in [2, 21] or run_copy()):
        logger.info("Skipping aggregate copy task. Not the 2nd or 21st of the month, and run_copy env variable is not true")
        return

    description = f"an `aggregate copy` job for projects: `{projects}` in namespaces: `{namespaces}`"
    if len(projects) == 0 and len(namespaces) == 0:
        description = "all root metadata"

    with insecure_channel(snapshot_address) as channel:
        stub = SnapshotsStub(channel)
        logger.info(f"Starting {description}")

        res = stub.AggregateCopy(
            AggregateCopyRequest(projects=projects, namespaces=namespaces),
            wait_for_ready=True
        )

        logger.info(
            f"Finished {description} "
            f"with total: `{res.total}` and errors: `{res.errors}`"
        )


def enable_snapshots_service():
    """
    Start tasks for snapshot ECS service
    """
    result = False
    ecs_cluster_name = getenv("ECS_CLUSTER_NAME")
    ecs_service_name = getenv("ECS_SNAPSHOTS_SERVICE_NAME")
    desired_count = get_desired_count()

    if any(not _ for _ in (ecs_cluster_name, ecs_service_name, desired_count)):
        raise AirflowException(
            "Refusing to proceed: mandatory environment variables are not set"
        )
    else:
        result = ecs_service_set_task_count(
            ecs_cluster_name, ecs_service_name, desired_count
        )

    if not result:
        raise AirflowException(f"Failure during service {ecs_service_name} update")

    return result


def disable_snapshots_service():
    """
    Stop tasks for snapshot ECS service
    """
    result = False
    ecs_cluster_name = getenv("ECS_CLUSTER_NAME")
    ecs_service_name = getenv("ECS_SNAPSHOTS_SERVICE_NAME")
    desired_count = 0

    if any(not _ for _ in (ecs_cluster_name, ecs_service_name)):
        raise AirflowException(
            "Refusing to proceed: mandatory environment variables are not set"
        )
    else:
        result = ecs_service_set_task_count(
            ecs_cluster_name, ecs_service_name, desired_count
        )

    if not result:
        raise AirflowException("We failed to update the service")

    return result


def verify_snapshots_service():
    """
    verify that DNS name of snapshot service resolves into desired count of IP addresses
    """
    desired_count = get_desired_count()
    result = verify_ip_addresses_returned_by_dns(addr_name, desired_count)

    if not result.success:
        raise AirflowException("Snapshot service verification has failed")

    return {"resolved_ips": result.ips}


def aggregate_chunks(projects, namespaces):
    """
    Call Snapshots service Aggregate API to generate aggregates of chunks of each snapshot.
    """
    if chunking_enabled:
        with insecure_channel(snapshot_address) as channel:
            for project in projects:
                for namespace in namespaces:
                    logger.info(
                        f"Starting a `chunk aggregation` job for project: '{project}' in namespace: '{namespace}'"
                    )
                    stub = SnapshotsStub(channel)
                    res = stub.Aggregate(
                        AggregateRequest(
                            prefix="chunks", snapshot=f"{project}_namespace_{namespace}"
                        ),
                        wait_for_ready=True
                    )

                    logger.info(
                        f"Finished a `chunk aggregation` job for project: `{project}` in namespace: "
                        f"`{namespace}` with total: `{res.total}` and errors: `{res.errors}`"
                    )

    else:
        logger.info("Chunking not enabled. Skipping aggregate chunk task.")

    return


with DAG(**dag_args) as dag:
    namespaces = get_namespaces(variable_name="namespaces_snapshots")
    projects = get_projects(variable_name="projects_snapshots")
    snapshots_total_articles = get_total_articles(variable_name="snapshots_total_articles")
    workers = get_copy_workers()
    exclude_events = get_exclude_events()

    copy_tasks = []

    for task_number, chunk in enumerate(get_chunks_by_project_size(snapshots_total_articles)):
        # opkwargs are used to provide parameters + context on top
        # https://stackoverflow.com/questions/67623479/use-kwargs-and-context-simultaniauly-on-airflow
        task_export = PythonOperator(
            task_id=f"export_{task_number}",
            python_callable=exports,
            op_kwargs=dict(
                projects=chunk,
                namespaces=namespaces,
                exclude_events=exclude_events,
                enable_chunking=chunking_enabled,
            ),
            execution_timeout=timedelta(hours=5),
            executor_config={
                "pod_template_file": pod_template_file
            }
        )

        task_copy = PythonOperator(
            task_id=f"copy_{task_number}",
            python_callable=copy,
            op_kwargs=dict(
                workers=workers,
                projects=chunk,
                namespaces=namespaces,
                datetime=datetime.now(),
            ),
            execution_timeout=timedelta(hours=2),
            executor_config={
                "pod_template_file": pod_template_file
            }
        )

        task_export >> task_copy
        copy_tasks.append(task_copy)

    task_aggregate = PythonOperator(
        task_id="aggregate",
        python_callable=aggregate,
        execution_timeout=timedelta(minutes=30),
        executor_config={
            "pod_template_file": pod_template_file
        }
    )

    # Copy the root metadata file, aggregations/snapshots/snapshots.ndjson
    task_aggregate_copy = PythonOperator(
        task_id="aggregate_copy",
        python_callable=aggregate_copy,
        op_kwargs=dict(
                projects=[],
                namespaces=[],
                datetime=datetime.now(),
        ),
        execution_timeout=timedelta(minutes=20),
        executor_config={
            "pod_template_file": pod_template_file
        }
    )

    task_aggregate_chunks = PythonOperator(
        task_id="aggregate_chunks",
        python_callable=aggregate_chunks,
        execution_timeout=timedelta(hours=3),
        executor_config={
            "pod_template_file": pod_template_file
        },
        op_kwargs=dict(
            projects=projects,
            namespaces=namespaces,
        ),
    )

    task_aggregate_chunks_copy = PythonOperator(
        task_id="aggregate_chunks_copy",
        python_callable=aggregate_copy,
        execution_timeout=timedelta(hours=1),
        executor_config={
            "pod_template_file": pod_template_file
        },
        op_kwargs=dict(
            projects=projects,
            namespaces=namespaces,
            datetime=datetime.now(),
        ),
    )

    copy_tasks >> task_aggregate >> task_aggregate_copy >> task_aggregate_chunks >> task_aggregate_chunks_copy
