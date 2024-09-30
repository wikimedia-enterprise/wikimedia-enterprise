"""
DAG for Commons Aggregations service. Contains AggregateCommons tasks
invoked on a monthly basis.
"""

from datetime import datetime, timedelta
from os import getenv
from airflow import DAG
from airflow.models import Variable
from airflow.operators.python import PythonOperator
from airflow.utils.trigger_rule import TriggerRule
from airflow import AirflowException
from logging import getLogger
from grpc import insecure_channel
from pipelines.utils.utils import (
    ecs_service_set_task_count,
    verify_ip_addresses_returned_by_dns,
    get_desired_count,
)

from pipelines.protos.snapshots_pb2_grpc import SnapshotsStub
from pipelines.protos.snapshots_pb2 import AggregateCommonsRequest

logger = getLogger("airflow.task")

addr_name = "AGGREGATIONS_SERVICE_ADDR"

dag_args = dict(
    dag_id="commons-aggregations",
    start_date=datetime(2024, 1, 1, 0, 0, 0),
    schedule_interval="@monthly",
    catchup=False,
    default_args=dict(provide_context=True),
)


def get_commons_batches_period():
    """
    Get commons_batches_period from Airflow Variable or calculate for previous month.
    """
    try:
        return Variable.get("commons_batches_period")
    except Exception:
        today = datetime.now()
        first_of_month = today.replace(day=1)
        last_month = first_of_month - timedelta(days=1)
        return last_month.strftime("%Y-%m")


def aggregate_commons(task_type, runner, verification_job_name, **kwargs):
    """
    Calls Aggregations service AggregateCommons API for batches or snapshots
    """
    if task_type == "batch" and not Variable.get("commons_batch_enabled", default_var=True, deserialize_json=True):
        logger.info("Commons batch aggregation is disabled. Skipping task.")
        return

    if task_type == "snapshot" and not Variable.get("commons_snapshot_enabled", default_var=True, deserialize_json=True):
        logger.info("Commons snapshot aggregation is disabled. Skipping task.")
        return

    shared_context = kwargs["ti"]
    ip_runners_pool = shared_context.xcom_pull(
        task_ids=verification_job_name
    ).get("resolved_ips")

    if not ip_runners_pool:
        raise AirflowException(
            f"aggregate_commons job is not able to start - we are lacking aggregation service IPs!: {ip_runners_pool}"
        )

    ip_address = ip_runners_pool[runner]

    logger.info(
        f"Taking {ip_address}(id {runner}) from {ip_runners_pool} to run the task"
    )

    with insecure_channel(ip_address) as channel:
        stub = SnapshotsStub(channel)
        
        if task_type == "batch":
            time_period = get_commons_batches_period()
            logger.info(f"Starting Commons batch aggregation for period: {time_period}")
            res = stub.AggregateCommons(
                AggregateCommonsRequest(
                    prefix="batches",
                    time_period=time_period
                )
            )
        elif task_type == "snapshot":
            logger.info("Starting Commons snapshot aggregation")
            res = stub.AggregateCommons(
                AggregateCommonsRequest(prefix="snapshots")
            )
        
        logger.info(
            f"Finished Commons {task_type} aggregation with total: {res.total} and errors: {res.errors}"
        )

    return


def enable_aggregations_service():
    """
    Start tasks for aggregations ECS service
    """
    result = False
    ecs_cluster_name = getenv("ECS_CLUSTER_NAME")
    ecs_service_name = getenv("ECS_AGGREGATIONS_SERVICE_NAME")
    desired_count = get_desired_count(env_variable_name="ECS_COMMONS_AGGREGATIONS_SERVICE_DESIRED_COUNT",
                                      airflow_variable_name="desired_count_commons_aggregations")

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


def disable_aggregations_service():
    """
    Stop tasks for aggregations ECS service
    """
    result = False
    ecs_cluster_name = getenv("ECS_CLUSTER_NAME")
    ecs_service_name = getenv("ECS_AGGREGATIONS_SERVICE_NAME")
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


def verify_aggregations_service():
    """
    Verify that DNS name of aggregations service resolves into desired count of IP addresses
    """
    desired_count = get_desired_count(env_variable_name="ECS_COMMONS_AGGREGATIONS_SERVICE_DESIRED_COUNT",
                                      airflow_variable_name="desired_count_commons_aggregations")
    result = verify_ip_addresses_returned_by_dns(addr_name, desired_count)

    if not result.success:
        raise AirflowException("Aggregations service verification has failed")

    return {"resolved_ips": result.ips}


with DAG(**dag_args) as dag:
    task_enable_aggregations = PythonOperator(
        task_id="enable_aggregations_service",
        python_callable=enable_aggregations_service,
        execution_timeout=timedelta(seconds=30),
    )

    task_verify_aggregations = PythonOperator(
        task_id="verify_aggregations_service",
        python_callable=verify_aggregations_service,
        execution_timeout=timedelta(seconds=960),
    )

    task_commons_batch = PythonOperator(
        task_id="commons_batch",
        python_callable=aggregate_commons,
        op_kwargs=dict(
            task_type="batch",
            runner=0,
            verification_job_name="verify_aggregations_service",
        ),
        execution_timeout=timedelta(days=2),
    )

    task_commons_snapshot = PythonOperator(
        task_id="commons_snapshot",
        python_callable=aggregate_commons,
        op_kwargs=dict(
            task_type="snapshot",
            runner=0,
            verification_job_name="verify_aggregations_service",
        ),
        execution_timeout=timedelta(days=2),
    )

    task_disable_aggregations = PythonOperator(
        task_id="disable_aggregations_service",
        python_callable=disable_aggregations_service,
        trigger_rule=TriggerRule.ALL_DONE,
        execution_timeout=timedelta(seconds=960),
    )

    task_verify_aggregations.set_upstream(task_enable_aggregations)
    task_commons_batch.set_upstream(task_verify_aggregations)
    task_commons_snapshot.set_upstream(task_commons_batch)
    task_disable_aggregations.set_upstream(task_commons_snapshot)
