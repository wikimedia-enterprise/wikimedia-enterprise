"""
DAG for bulk injection service. Contains Project, Namespace and Arctiles tasks
each invoked on the manual basis.
"""
import sys
from os import getenv, path
from datetime import timedelta, datetime
from airflow import DAG, AirflowException
from airflow.operators.python import PythonOperator
from airflow.providers.cncf.kubernetes.hooks.kubernetes import KubernetesHook
from airflow.utils.task_group import TaskGroup
from airflow.models import Variable
from grpc import insecure_channel
from airflow.settings import AIRFLOW_HOME
import kubernetes.client
from kubernetes.client.rest import ApiException
from time import sleep 
DAGS_HOME=path.join(AIRFLOW_HOME, "dags/git_wme-dags")
WORKERS_HOME=path.join(AIRFLOW_HOME, "dags/dags.git")

sys.path += [DAGS_HOME, WORKERS_HOME]
print(sys.path)
from pipelines.utils.utils import (
    get_namespaces,
    get_projects,
    get_chunks_by_project_size,
    get_total_articles,
)
from logging import getLogger
from kafka import KafkaAdminClient, KafkaConsumer
from time import sleep
from json import loads

from pipelines.protos.bulk_pb2_grpc import BulkStub
from pipelines.protos.bulk_pb2 import (
    ProjectsRequest,
    NamespacesRequest,
    ArticlesRequest,
)

bulk_ingestion_address = '0.0.0.0:5050'
environment  = getenv('CLUSTER_ENV')
pod_template_file = path.join(DAGS_HOME, f'pod_templates/bulk-ingestion-{environment}.yaml')
generic_pod_template_file = path.join(DAGS_HOME, f'pod_templates/bulk-generic-{environment}.yaml')
DESIRED_BULK_REPLICAS = int(Variable.get("article_bulk_replicas", default_var=2))
DESIRED_BULK_ERROR_REPLICAS= int(Variable.get("article_bulk_error_replicas", default_var=1))
ARTICLE_BULK_NAMESPACE = "structured-data"
K8S_CONN_ID = "kubernetes_default"
SLEEP_INTERVAL = 20
MAX_RETRIES = 30

# Initialize logger
logger = getLogger("airflow.task")

def scale_article_bulk_service(deployment_name: str, replicas: int, **kwargs):
    """Scales article-bulk deployment and validates the number of running replicas."""
    try:
        k8s_hook = KubernetesHook(conn_id=K8S_CONN_ID)
        api_client = k8s_hook.get_conn()
        api_instance = kubernetes.client.AppsV1Api(api_client)

        # Scale deployment
        body = {"spec": {"replicas": replicas}}
        api_instance.patch_namespaced_deployment_scale(
            name=deployment_name, namespace=ARTICLE_BULK_NAMESPACE, body=body
        )
        logger.info(f"Scaling deployment {deployment_name} to {replicas} replicas...")

        # Validate replicas are running
        for attempt in range(MAX_RETRIES):
            response = api_instance.read_namespaced_deployment(
                name=deployment_name, namespace=ARTICLE_BULK_NAMESPACE
            )
            ready_replicas = response.status.ready_replicas or 0
            if ready_replicas == replicas:
                print(f"Desired replicas: {ready_replicas} achieved.")
                return
            logger.info(f"Attempt {attempt + 1}: Waiting for the desired replicas to be available...")
            sleep(SLEEP_INTERVAL)

        raise AirflowException(
            f"Validation failed: Expected {replicas} replicas, but found {ready_replicas}."
        )
    except ApiException as e:
        raise AirflowException(f"Kubernetes API error: {e}")

    except OSError as e:  # Handles missing kubeconfig or connection failure
        raise AirflowException(f"Failed to connect to Kubernetes cluster: {e}")

    except Exception as e:
        raise AirflowException(f"Unexpected error while initializing KubernetesHook: {e}")



def articles(projects, namespaces):
    """
    Call Bulk-Ingestion service API to get all the article titles by
    calling mediawiki all pages API.
    Then, it splits pages to chunks of max 50 article names and returns
    produces a kafka message for each chunk.
    """
    logger.info(f"Got namespaces configuration: `{namespaces}`\n")
    logger.info(f"Got projects configuration: `{projects}`\n")

    for project in projects:
        for namespace in namespaces:
            with insecure_channel(bulk_ingestion_address) as channel:
                logger.info(
                    f"Starting an `articles` job for project: '{project}' in namespace: '{namespace}'\n"
                )

                res = BulkStub(channel).Articles(
                    ArticlesRequest(
                        project=project,
                        namespace=namespace,
                    ),
                    wait_for_ready=True
                )

                logger.info(
                    f"Finished an `articles` job for project: `{project}` "
                    f"in namespace: `{namespace}` with total: `{res.total}` and errors: `{res.errors}`\n"
                )

    return "Articles processing completed"


def projects():
    """
    Call Bulk-ingestion service API to get all the projects by
    calling mediawiki actions API, and produces a kafka message
    for each project.
    """
    with insecure_channel(bulk_ingestion_address) as channel:
        logger.info("Starting the `projects` job")

        stub = BulkStub(channel)
        res = stub.Projects(
            ProjectsRequest(),
            wait_for_ready=True
            )

        logger.info(
            f"Finished the `projects` job with total: `{res.total}` and errors: `{res.errors}`\n"
        )
    return


def namespaces():
    """
    Call Bulk-ingestion service API to produce a kafka message for each
    namespace per project.
    """
    with insecure_channel(bulk_ingestion_address) as channel:
        logger.info("Starting the `namespaces` job")

        stub = BulkStub(channel)
        res = stub.Namespaces(
            NamespacesRequest(),
            wait_for_ready=True
            )

        logger.info(
            f"Finished the `namespaces` job with total: `{res.total}` and errors: `{res.errors}`\n"
        )
    return


def calculate_consumption_lag(kafka_admin, kafka_consumer, group_id):
    """
    Calculate the lag of the consumer group.

    Returns:
        int: consumer lag
    """
    topic_partitions = []
    consumer_offsets = {}
    offsets = kafka_admin.list_consumer_group_offsets(group_id=group_id)

    for topic_partition, meta in offsets.items():
        consumer_offsets[topic_partition] = meta.offset
        topic_partitions.append(topic_partition)

    end_offsets = kafka_consumer.end_offsets(topic_partitions)
    consumer_lag = 0

    for topic_partition, offset in end_offsets.items():
        consumer_lag += offset - consumer_offsets[topic_partition]

    return consumer_lag


def get_consumer_group_state(kafka_admin, group_id):
    """
    Get the state of the consumer group.

    Returns:
        str: consumer group state ('Stable', 'Empty', 'Dead', 'PreparingRebalance', 'CompletingRebalance')
    """
    consumer_groups = kafka_admin.describe_consumer_groups(group_ids=[group_id])
    if len(consumer_groups) > 0:
        return consumer_groups[0].state

    return ""


def monitor_article_bulk_service():
    """
    Monitor article bulk service to see how much data was processed and still needs to be processed.
    The monitoring is happening using consumer group `messages behind` property.
    """
    # credentials_raw = get_secret_value(getenv("KAFKA_CREDS_SECRET_NAME"))
    credentials_raw = getenv("KAFKA_CREDS")

    if len(credentials_raw) == 0:
        raise AirflowException(
            "Refusing to proceed: mandatory environment variable `KAFKA_CREDS_SECRET_NAME` is not set"
        )

    kafka_bootstrap_servers = getenv("KAFKA_BOOTSTRAP_SERVERS")

    if len(kafka_bootstrap_servers) == 0:
        raise AirflowException(
            "Refusing to proceed: mandatory environment variable `KAFKA_BOOTSTRAP_SERVERS` is not set"
        )

    credentials = loads(credentials_raw)
    kafka_admin = KafkaAdminClient(
        bootstrap_servers=kafka_bootstrap_servers,
        security_protocol="SASL_SSL",
        sasl_mechanism="SCRAM-SHA-512",
        sasl_plain_username=credentials["username"],
        sasl_plain_password=credentials["password"],
    )
    kafka_consumer = KafkaConsumer(
        bootstrap_servers=kafka_bootstrap_servers,
        security_protocol="SASL_SSL",
        sasl_mechanism="SCRAM-SHA-512",
        sasl_plain_username=credentials["username"],
        sasl_plain_password=credentials["password"],
    )

    # Kafka consumer group id
    group_id = Variable.get(
        "articlebulk_consumer_group", default_var="structured-data-articlebulk-v2"
    )

    while True:
        state = get_consumer_group_state(kafka_admin, group_id)

        logger.info(f"Consumer group `{group_id}` is in `{state}` state")

        if state == "Stable":
            logger.info(
                f"Consumer group `{group_id}` reached a `{state}` state, continuing with the task"
            )
            break

        logger.info(
            "Group is not `Stable`, not ready to proceed, waiting for 15 seconds before retrying"
        )
        sleep(15)

    while True:
        consumer_lag = calculate_consumption_lag(kafka_admin, kafka_consumer, group_id)

        logger.info(f"Consumer lag is `{consumer_lag}`")

        if consumer_lag == 0:
            logger.info("Zero consumer lag reached, continuing with the task")
            break

        logger.info("Consumer lag is not zero, waiting for 15 seconds before retrying")
        sleep(15)
    return


with DAG(
    dag_id="bulk-ingestion",
    start_date=datetime(2022, 1, 1, 0, 0, 0),
    schedule_interval=None,
) as dag:
    task_scale_up_articlebulk = PythonOperator(
        task_id="scale_up_article_bulk",
        python_callable=scale_article_bulk_service,
        op_kwargs=dict(
            deployment_name="structured-data-articlebulk-live",
            replicas=DESIRED_BULK_REPLICAS,
        ),
        execution_timeout=timedelta(days=30),
        executor_config={
            "pod_template_file": generic_pod_template_file
        }
    )

    task_scale_up_articlebulk_error = PythonOperator(
        task_id="scale_up_article_bulk_error",
        python_callable=scale_article_bulk_service,
        op_kwargs=dict(
            deployment_name="structured-data-articlebulk-error-live",
            replicas=DESIRED_BULK_ERROR_REPLICAS,
        ),
        execution_timeout=timedelta(days=30),
        executor_config={
            "pod_template_file": generic_pod_template_file
        }
    )


    task_namespaces = PythonOperator(
        task_id="namespaces",
        python_callable=namespaces,
        execution_timeout=timedelta(hours=1),
        executor_config={
            "pod_template_file": pod_template_file
        }
    )

    task_projects = PythonOperator(
        task_id="projects",
        python_callable=projects,
        execution_timeout=timedelta(hours=1),
        executor_config={
            "pod_template_file": pod_template_file
        }
    )

    task_monitor_article_bulk = PythonOperator(
        task_id="monitor_article_bulk",
        python_callable=monitor_article_bulk_service,
        execution_timeout=timedelta(days=30),
        executor_config={
            "pod_template_file": generic_pod_template_file
        }
    )

    [task_scale_up_articlebulk, task_scale_up_articlebulk_error] >> task_monitor_article_bulk
    

    conf_namespaces = get_namespaces(variable_name="namespaces_bulk_ingestion")
    conf_projects = get_projects(variable_name="projects_bulk_ingestion")

    # # Get the total articles for all project in a CSV file
    article_counts = get_total_articles(variable_name="snapshots_total_articles")
    filtered_groups = {k: article_counts[k] for k in conf_projects if k in article_counts}
    print("Filtered Groups:", filtered_groups)

    grouped_projects = get_chunks_by_project_size(filtered_groups)
    print("Grouped Projects:", grouped_projects)

    for i, projects in enumerate(grouped_projects):
        task_articles = PythonOperator(
            task_id=f"articles_{i}",
            python_callable=articles,
            op_kwargs={
                "projects": projects,
                "namespaces": conf_namespaces
            },
            execution_timeout=timedelta(days=10),
            executor_config={"pod_template_file": pod_template_file},
        )
    
    task_scale_down_articlebulk = PythonOperator(
        task_id="scale_down_article_bulk",
        python_callable=scale_article_bulk_service,
        op_kwargs=dict(
            deployment_name="structured-data-articlebulk-live",
            replicas=0,
        ),
        execution_timeout=timedelta(days=30),
        executor_config={
            "pod_template_file": generic_pod_template_file
        }
    )

    task_scale_down_articlebulk_error = PythonOperator(
        task_id="scale_down_article_bulk_error",
        python_callable=scale_article_bulk_service,
        op_kwargs=dict(
            deployment_name="structured-data-articlebulk-error-live",
            replicas=0,
        ),
        execution_timeout=timedelta(days=30),
        executor_config={
            "pod_template_file": generic_pod_template_file
        }
    )

    task_monitor_article_bulk >> [task_scale_down_articlebulk, task_scale_down_articlebulk_error]

 


