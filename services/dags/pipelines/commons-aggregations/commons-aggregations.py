"""
DAG for Commons Aggregations service. Contains AggregateCommons tasks
invoked on a monthly basis.
"""

from datetime import datetime, timedelta
from os import getenv, path
import sys
from airflow import DAG
from airflow.models import Variable
from airflow.operators.python import PythonOperator
from airflow.utils.trigger_rule import TriggerRule
from airflow import AirflowException
from logging import getLogger
from grpc import insecure_channel
from airflow.settings import AIRFLOW_HOME


DAGS_HOME= path.join(AIRFLOW_HOME, "dags/git_wme-dags")
WORKERS_HOME= path.join(AIRFLOW_HOME, "dags/dags.git")
sys.path += [DAGS_HOME, WORKERS_HOME]
# from pipelines.utils.utils import (
#     get_list_iterator,
#     verify_ip_addresses_returned_by_dns,
#     get_desired_count,
# )

from pipelines.protos.snapshots_pb2_grpc import SnapshotsStub
from pipelines.protos.snapshots_pb2 import AggregateCommonsRequest


environment  = getenv('CLUSTER_ENV')
pod_template_file = path.join(DAGS_HOME, f'pod_templates/aggregation-{environment}.yaml')

logger = getLogger("airflow.task")

service_address = "0.0.0.0:5051"

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


def aggregate_commons(task_type, **kwargs):
    """
    Calls Aggregations service AggregateCommons API for batches or snapshots
    """
    if task_type == "batch" and not Variable.get("commons_batch_enabled", default_var=True, deserialize_json=True):
        logger.info("Commons batch aggregation is disabled. Skipping task.")
        return

    if task_type == "snapshot" and not Variable.get("commons_snapshot_enabled", default_var=True, deserialize_json=True):
        logger.info("Commons snapshot aggregation is disabled. Skipping task.")
        return
   

    with insecure_channel(service_address) as channel:
        stub = SnapshotsStub(channel)
        
        if task_type == "batch":
            time_period = get_commons_batches_period()
            logger.info(f"Starting Commons batch aggregation for period: {time_period}")
            res = stub.AggregateCommons(
                AggregateCommonsRequest(
                    prefix="batches",
                    time_period=time_period
                ),
                wait_for_ready=True
            )
        elif task_type == "snapshot":
            logger.info("Starting Commons snapshot aggregation")
            res = stub.AggregateCommons(
                AggregateCommonsRequest(
                    prefix="snapshots"
                ),
                wait_for_ready=True
            )
        
        logger.info(
            f"Finished Commons {task_type} aggregation with total: {res.total} and errors: {res.errors}"
        )

    return


with DAG(**dag_args) as dag:

    task_commons_batch = PythonOperator(
        task_id="commons_batch",
        python_callable=aggregate_commons,
        op_kwargs=dict(
            task_type="batch",
            verification_job_name="verify_aggregations_service",
        ),
        execution_timeout=timedelta(days=2),
        executor_config = {
          "pod_template_file": pod_template_file 
        }
    )

    task_commons_snapshot = PythonOperator(
        task_id="commons_snapshot",
        python_callable=aggregate_commons,
        op_kwargs=dict(
            task_type="snapshot",
            verification_job_name="verify_aggregations_service",
        ),
        execution_timeout=timedelta(days=4),
        executor_config = {
          "pod_template_file": pod_template_file 
        }
    )

