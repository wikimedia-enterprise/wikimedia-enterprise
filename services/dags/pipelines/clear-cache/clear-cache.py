"""
DAG to connect to an elasticache instance. It deletes all the keys that follow any of 
the patterns provided in an airflow variable `cache_patterns`. 
Provide a list of glob-style patterns (https://redis.io/docs/latest/commands/keys/)
For example with ["cap:*"]
- key examples deleted by this: cap:ondemand:user:user-xyz:count, cap:snapshot:user:user-x:count, etc.
"""

from os import getenv,path
from logging import getLogger
from airflow import DAG
from airflow.operators.python_operator import PythonOperator
from airflow.models import Variable
from airflow.settings import AIRFLOW_HOME
from datetime import datetime, timedelta
import redis
import sys

logger = getLogger("airflow.task")
environment = getenv('CLUSTER_ENV')
DAGS_HOME=path.join(AIRFLOW_HOME, "dags/git_wme-dags")
WORKERS_HOME=path.join(AIRFLOW_HOME, "dags/dags.git")
sys.path += [DAGS_HOME, WORKERS_HOME]
pod_template_file = path.join(DAGS_HOME, f'pod_templates/generic-{environment}.yaml')
REDIS_PORT = 6379

def clear_keys():
    cache_patterns = Variable.get(
        "cache_patterns", default_var=["cap:*"], deserialize_json=True
    )
    host = getenv("REDIS_ADDR")
    port = REDIS_PORT

    try:
        client = redis.Redis(
            host=host,
            port=port,
            password=getenv("REDIS_PASSWORD"),
            ssl=True,  
            decode_responses=True,
            socket_timeout=10,
            socket_connect_timeout=10
            )
        
        client.ping()
    except redis.RedisError as e:
        logger.error(f"Error connecting to elasticache: {e}")
        return

    logger.info(f"Successfully connected to elasticache {host}:{port}")
    logger.info(f"About to delete all keys matching patterns '{cache_patterns}'")

    total = 0

    # Iterate over each pattern and delete matching keys
    for pattern in cache_patterns:
        try:
            matched_keys = []
            cursor = 0
            while cursor != 0 or not matched_keys:
                cursor, keys = client.scan(cursor=cursor, match=pattern, count=10)
                matched_keys.extend(keys)
                if cursor == 0:
                    break  # Stop scanning
        except Exception as e:
            logger.error(f"Error retrieving keys for pattern '{pattern}': {e}")
            return

        logger.info(f"For pattern: {pattern}, \nFound matched keys: {matched_keys}")
        
        if not matched_keys:
            logger.info(f"No keys found matching pattern '{pattern}'")
            continue

        for key in matched_keys:
            try:
                client.delete(key)
            except Exception as e:
                logger.error(f"Error deleting key {key}: {e}")
                return

            logger.info(f"Successfully deleted key {key}")
            total += 1

    logger.info(f"Deleted {total} keys from cache")


dag_args = dict(
    dag_id="clear_cache",
    start_date=datetime(2024, 1, 1, 0, 0, 0),
    schedule_interval="@monthly",
    catchup=False,
    default_args=dict(provide_context=True),
)

with DAG(**dag_args) as dag:
    task_clear_keys = PythonOperator(
        task_id="clear_keys",
        python_callable=clear_keys,
        execution_timeout=timedelta(minutes=30),
        executor_config={
            "pod_template_file": pod_template_file
        }
    )
