"""
DAG for Snapshots services. Contains Export and Aggregate tasks
each invoked on an daily basis.
"""
from os import getenv
from datetime import timedelta, datetime
from airflow import DAG
from airflow.models import Variable
from airflow.operators.python import PythonOperator
from airflow.utils.trigger_rule import TriggerRule
from airflow import AirflowException
from logging import getLogger
from grpc import insecure_channel
from pipelines.utils.utils import (
    get_channel,
    get_list_iterator,
    get_namespaces,
    get_projects,
    get_language,
    get_batch_size,
    get_copy_workers,
    run_copy,
    get_exclude_events,
    chunks,
    ecs_service_set_task_count,
    verify_ip_addresses_returned_by_dns,
    get_desired_count,
)

from pipelines.protos.snapshots_pb2_grpc import SnapshotsStub
from pipelines.protos.snapshots_pb2 import (
    ExportRequest,
    CopyRequest,
    AggregateRequest,
    AggregateCopyRequest,
)

# Initialize logger
logger = getLogger("airflow.task")

# Initialize service env variable name
addr_name = "SNAPSHOTS_SERVICE_ADDR"

dag_args = dict(
    dag_id="snapshots",
    start_date=datetime(2022, 1, 1, 0, 0, 0),
    schedule_interval="@daily",
    catchup=False,
    default_args=dict(provide_context=True),
)


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
    projects, namespaces, exclude_events, runner, verification_job_name, **kwargs
):
    """
    Calls Snapshots service Export API for each project-ns
    """

    # get shared context
    shared_shared_context = kwargs["ti"]

    # pull value from xcom buffer, from the shared context
    ip_runners_pool = shared_shared_context.xcom_pull(
        task_ids=verification_job_name
    ).get("resolved_ips")

    if not ip_runners_pool:
        raise AirflowException(
            f"export job is not able ti start - we are lacking snapshot service IPs!: {ip_runners_pool}"
        )

    # get an IP address by an "id" of the runner
    ip_address = ip_runners_pool[runner]

    logger.info(
        f"Taking {ip_address}(id {runner}) from {ip_runners_pool} to run the task"
    )

    logger.info(f"Got namespaces configuration: `{namespaces}`")
    logger.info(f"Got projects configuration: `{projects}`")
    logger.info(f"Exclude events of type `{exclude_events}`")
    logger.info(f"Trying to get channel on ip address `{ip_address}`")

    since = get_since()

    if since > 0:
        logger.info(
            f"Exporting snapshots with since parameter set to `{since}`")
    else:
        logger.info(
            "Exporting snapshots without since parameter, all articles will be exported"
        )

    with insecure_channel(ip_address) as channel:
        for project in projects:
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
                    )
                )

                logger.info(
                    f"Finished an `export` job for project: `{project}` in namespace: "
                    f"`{namespace}` with total: `{res.total}` and errors: `{res.errors}`"
                )
    return


def copy(
    workers, projects, namespaces, datetime, runner, verification_job_name, **kwargs
):
    """
    Calls Snapshots service Copy API for a ns, and a set of projects.
    """
    if datetime.day == 1 or run_copy():
        # get shared context
        shared_shared_context = kwargs["ti"]

        # pull value from xcom buffer, from the shared context
        ip_runners_pool = shared_shared_context.xcom_pull(
            task_ids=verification_job_name
        ).get("resolved_ips")

        if not ip_runners_pool:
            raise AirflowException(
                f"copy job is not able to start - we are lacking snapshot service IPs!: {ip_runners_pool}"
            )

        # get an IP address by an "id" of the runner
        ip_address = ip_runners_pool[runner]

        logger.info(
            f"Taking {ip_address}(id {runner}) from {ip_runners_pool} to run the task"
        )

        logger.info(f"Got namespaces configuration: `{namespaces}`")
        logger.info(f"Got projects configuration: `{projects}`")
        logger.info(f"Got workers configuration: `{workers}`")
        logger.info(f"Trying to get channel on ip address `{ip_address}`")

        with insecure_channel(ip_address) as channel:
            for namespace in namespaces:
                logger.info(
                    f"Starting a `copy` job for namespace: '{namespace}' for projects '{projects}'"
                )

                res = SnapshotsStub(channel).Copy(
                    CopyRequest(workers=workers, projects=projects,
                                namespace=namespace)
                )

                logger.info(
                    f"Finished a `copy` job for projects: `{projects}` in namespace: "
                    f"`{namespace}` with total: `{res.total}` and errors: `{res.errors}`"
                )
    else:
        logger.info("Not the first of the month. Skipping copy task.")

    return


def aggregate():
    """
    Call Snapshots service Aggregate API to generate download links
    for each export.
    """
    with get_channel(addr_name) as channel:
        logger.info("Starting the `aggregate` job")

        stub = SnapshotsStub(channel)
        res = stub.Aggregate(AggregateRequest())

        logger.info(
            "Finished the `aggregate` job with total: `{total}` and errors: `{errors}`\n".format(
                total=res.total,
                errors=res.errors,
            )
        )

    return


def aggregate_copy(datetime):
    """
    Call Snapshots service AggregateCopy API to copy aggregated metadata of snapshots.
    """
    if datetime.day == 1 or run_copy():
        with get_channel(addr_name) as channel:
            logger.info("Starting the `aggregate copy` job")

            stub = SnapshotsStub(channel)
            res = stub.AggregateCopy(AggregateCopyRequest())

            logger.info(
                "Finished the `aggregate copy` job with total: `{total}` and errors: `{errors}`\n".format(
                    total=res.total,
                    errors=res.errors,
                )
            )
    else:
        logger.info("Not the first of the month. Skipping aggregate copy task.")

    return


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
        raise AirflowException(
            f"Failure during service {ecs_service_name} update")

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


with DAG(**dag_args) as dag:
    task_aggregate = PythonOperator(
        task_id="aggregate",
        python_callable=aggregate,
        execution_timeout=timedelta(days=2),
    )

    task_aggregate_copy = PythonOperator(
        task_id="aggregate_copy",
        python_callable=aggregate_copy,
        op_kwargs=dict(datetime=datetime.now()),
        execution_timeout=timedelta(minutes=10),
    )

    # stop the ECS service in any circumstances, as the last resort step
    task_disable_snapshots = PythonOperator(
        task_id="disable_snapshots_service",
        python_callable=disable_snapshots_service,
        trigger_rule=TriggerRule.ALL_DONE,
        execution_timeout=timedelta(seconds=30),
    )
    # set order after the aggregation job
    task_disable_snapshots.set_upstream(task_aggregate_copy)

    # enable service in ECS
    task_enable_snapshots = PythonOperator(
        task_id="enable_snapshots_service",
        python_callable=enable_snapshots_service,
        execution_timeout=timedelta(seconds=30),
    )

    # verify that desired amount of services are available
    task_verify_snapshots = PythonOperator(
        task_id="verify_snapshots_service",
        python_callable=verify_snapshots_service,
        execution_timeout=timedelta(seconds=960),
    )

    # ensure verification runs only after starting the snapshot service
    task_verify_snapshots.set_upstream(task_enable_snapshots)

    # generate dummy list with runner IDs, to distribute the jobs across the runners,
    # since we are awaiting desired amount of services - we
    # can assign runner IDs beforehand, and later
    desired_count = get_desired_count()
    runners_ids = get_list_iterator(list(range(desired_count)))

    namespaces = get_namespaces(variable_name="namespaces_snapshots")
    projects = get_projects(variable_name="projects_snapshots")
    workers = get_copy_workers()
    exclude_events = get_exclude_events()

    for task_number, chunk in enumerate(chunks(projects, get_batch_size(variable_name="batch_size_snapshots"))):
        # opkwargs are used to provide parameters + context on top
        # https://stackoverflow.com/questions/67623479/use-kwargs-and-context-simultaniauly-on-airflow
        task_export = PythonOperator(
            task_id=f"export_{task_number}",
            python_callable=exports,
            op_kwargs=dict(
                projects=chunk,
                namespaces=namespaces,
                exclude_events=exclude_events,
                runner=next(runners_ids),
                verification_job_name="verify_snapshots_service",
            ),
            execution_timeout=timedelta(days=7),
        )

        task_copy = PythonOperator(
            task_id=f"copy_{task_number}",
            python_callable=copy,
            op_kwargs=dict(
                workers=workers,
                projects=chunk,
                namespaces=namespaces,
                datetime=datetime.now(),
                runner=next(runners_ids),
                verification_job_name="verify_snapshots_service",
            ),
            execution_timeout=timedelta(hours=2),
        )

        # ensure verification is successful before triggering any export job
        task_export.set_upstream(task_verify_snapshots)

        task_copy.set_upstream(task_export)

        # ensure all exports are finished before running the aggregation
        task_aggregate.set_upstream(task_copy)

    task_aggregate_copy.set_upstream(task_aggregate)
