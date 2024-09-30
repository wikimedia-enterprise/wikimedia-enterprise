"""
DAG for bulk injection service. Contains Project, Namespace and Arctiles tasks
each invoked on the manual basis.
"""

from datetime import timedelta, datetime
from airflow import DAG, AirflowException
from airflow.operators.python import PythonOperator
from airflow.models import Variable
from pipelines.utils.utils import (
    get_secret_value,
    get_channel,
    get_namespaces,
    get_projects,
    get_desired_count,
    ecs_service_set_task_count,
    verify_ip_addresses_returned_by_dns,
)
from logging import getLogger
from os import getenv
from kafka import KafkaAdminClient, KafkaConsumer
from time import sleep
from json import loads

from pipelines.protos.bulk_pb2_grpc import BulkStub
from pipelines.protos.bulk_pb2 import (
    ProjectsRequest,
    NamespacesRequest,
    ArticlesRequest,
)

# Initialize logger
logger = getLogger("airflow.task")

# Initialize service env variable name
addr_name = "BULK_INGESTION_SERVICE_ADDR"


def get_bulk_ingestion_channel():
    """
    Wrapper function to simplify getting the bulk ingestion channel.

    Returns:
        grpc.Channel: channel to the bulk ingestion service
    """
    return get_channel(addr_name)


def articles(projects, namespaces):
    """
    Call Bulk-Ingestion service API to get all the article titles by
    calling mediawiki all pages API.
    Then, it splits pages to chunks of max 50 article names and returns
    produces a kafka message for each chunk.
    """

    def articles():
        logger.info(f"Got namespaces configuration: `{namespaces}`\n")
        logger.info(f"Got projects configuration: `{projects}`\n")

        for project in projects:
            for namespace in namespaces:
                with get_bulk_ingestion_channel() as channel:
                    logger.info(
                        f"Starting an `articles` job for project: '{project}' in namespace: '{namespace}'\n"
                    )

                    res = BulkStub(channel).Articles(
                        ArticlesRequest(
                            project=project,
                            namespace=namespace,
                        )
                    )

                    logger.info(
                        f"Finished an `articles` job for project: `{project}` \
                                in namespace: `{namespace}` with total: `{res.total}` and errors: `{res.errors}`\n"
                    )
        return

    return articles


def projects():
    """
    Call Bulk-ingestion service API to get all the projects by
    calling mediawiki actions API, and produces a kafka message
    for each project.
    """
    with get_bulk_ingestion_channel() as channel:
        logger.info("Starting the `projects` job")

        stub = BulkStub(channel)
        res = stub.Projects(ProjectsRequest())

        logger.info(
            f"Finished the `projects` job with total: `{res.total}` and errors: `{res.errors}`\n"
        )
    return


def namespaces():
    """
    Call Bulk-ingestion service API to produce a kafka message for each
    namespace per project.
    """
    with get_bulk_ingestion_channel() as channel:
        logger.info("Starting the `namespaces` job")

        stub = BulkStub(channel)
        res = stub.Namespaces(NamespacesRequest())

        logger.info(
            f"Finished the `namespaces` job with total: `{res.total}` and errors: `{res.errors}`\n"
        )
    return


def get_bulk_ingestion_desired_count():
    """
    Wrapper function to simplify getting the desired count for the bulk ingestion service.

    Returns:
        int: number of desired replicas for the bulk ingestion service
    """
    return get_desired_count(
        env_variable_name="ECS_BULK_INGESTION_SERVICE_DESIRED_COUNT",
        airflow_variable_name="desired_count_bulk_ingestion",
        default_desired_count=1,
    )


def get_bulk_ingestion_ecs_service_name():
    """
    Function to simplify the lookup of the bulk ingestion service name.

    Returns:
        str: name of the bulk ingestion service
    """
    return getenv("ECS_BULK_INGESTION_SERVICE_NAME")


def get_ecs_cluster_name():
    """
    Function to simplify the lookup of the ECS cluster name.

    Returns:
        str: name of the ECS cluster
    """
    return getenv("ECS_CLUSTER_NAME")


def enable_bulk_ingestion_service():
    """
    Upscale bulk ingestion service to desired amounts of replicas.
    Will get the number of replicas from the airflow variable or env variable.
    """
    ecs_cluster_name = get_ecs_cluster_name()
    ecs_service_name = get_bulk_ingestion_ecs_service_name()
    desired_count = get_bulk_ingestion_desired_count()

    if any(not _ for _ in (ecs_cluster_name, ecs_service_name, desired_count)):
        raise AirflowException(
            "Refusing to proceed: mandatory environment variables are not set"
        )

    result = ecs_service_set_task_count(
        ecs_cluster_name, ecs_service_name, desired_count
    )

    if not result:
        raise AirflowException(f"Failure during service {ecs_service_name} update")
    return


def verify_bulk_ingestion_service():
    """
    Verify that the bulk ingestion service was really up-scaled.
    """
    desired_count = get_bulk_ingestion_desired_count()
    result = verify_ip_addresses_returned_by_dns(addr_name, desired_count)

    if not result.success:
        raise AirflowException("Service verification has failed")
    return


def disable_bulk_ingestion_service():
    """
    Downscale bulk ingestion service to 0 replicas.
    """
    ecs_cluster_name = get_ecs_cluster_name()
    ecs_service_name = get_bulk_ingestion_ecs_service_name()

    if any(not _ for _ in (ecs_cluster_name, ecs_service_name)):
        raise AirflowException(
            "Refusing to proceed: mandatory environment variables are not set"
        )

    result = ecs_service_set_task_count(ecs_cluster_name, ecs_service_name, 0)

    if not result:
        raise AirflowException(f"Failure during service {ecs_service_name} update")
    return


def get_article_bulk_ecs_service_name():
    """
    Function to simplify the lookup of the article bulk service name.

    Returns:
        str: name of the article bulk service
    """
    return getenv("ECS_ARTICLE_BULK_SERVICE_NAME")


def get_article_bulk_desired_count():
    """
    Wrapper function to simplify getting the desired count for the article bulk service.

    Returns:
        int: number of desired replicas for the article bulk service
    """
    return get_desired_count(
        env_variable_name="ECS_ARTICLE_BULK_SERVICE_DESIRED_COUNT",
        airflow_variable_name="desired_count_article_bulk",
        default_desired_count=5,
    )


def enable_article_bulk_service():
    """
    Upscale article bulk service to desired amounts of replicas.
    Will get the number of replicas from the airflow variable or env variable.
    """
    ecs_cluster_name = get_ecs_cluster_name()
    ecs_service_name = get_article_bulk_ecs_service_name()
    desired_count = get_article_bulk_desired_count()

    if any(not _ for _ in (ecs_cluster_name, ecs_service_name, desired_count)):
        raise AirflowException(
            "Refusing to proceed: mandatory environment variables are not set"
        )

    result = ecs_service_set_task_count(
        ecs_cluster_name, ecs_service_name, desired_count
    )

    if not result:
        raise AirflowException(f"Failure during service {ecs_service_name} update")
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
    credentials_raw = get_secret_value(getenv("KAFKA_CREDS_SECRET_NAME"))

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
        "consumer_group_id", default_var="structured-data-articlebulk-v2"
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


def disable_article_bulk_service():
    """
    Downscale article bulk service to 0 replicas.
    """
    ecs_cluster_name = get_ecs_cluster_name()
    ecs_service_name = get_article_bulk_ecs_service_name()

    if any(not _ for _ in (ecs_cluster_name, ecs_service_name)):
        raise AirflowException(
            "Refusing to proceed: mandatory environment variables are not set"
        )

    result = ecs_service_set_task_count(ecs_cluster_name, ecs_service_name, 0)

    if not result:
        raise AirflowException(f"Failure during service {ecs_service_name} update")
    return


with DAG(
    dag_id="bulk-ingestion",
    start_date=datetime(2022, 1, 1, 0, 0, 0),
    schedule_interval=None,
) as dag:
    task_enable_bulk_ingestion = PythonOperator(
        task_id="enable_bulk_ingestion",
        python_callable=enable_bulk_ingestion_service,
        execution_timeout=timedelta(seconds=30),
    )

    task_verify_bulk_ingestion = PythonOperator(
        task_id="verify_bulk_ingestion",
        python_callable=verify_bulk_ingestion_service,
        execution_timeout=timedelta(seconds=960),
    )

    task_verify_bulk_ingestion.set_upstream(task_enable_bulk_ingestion)

    task_disable_bulk_ingestion = PythonOperator(
        task_id="disable_bulk_ingestion",
        python_callable=disable_bulk_ingestion_service,
        execution_timeout=timedelta(seconds=30),
    )

    task_namespaces = PythonOperator(
        task_id="namespaces",
        python_callable=namespaces,
        execution_timeout=timedelta(hours=1),
    )

    task_namespaces.set_upstream(task_verify_bulk_ingestion)
    task_namespaces.set_downstream(task_disable_bulk_ingestion)

    task_projects = PythonOperator(
        task_id="projects",
        python_callable=projects,
        execution_timeout=timedelta(hours=1),
    )

    task_projects.set_upstream(task_verify_bulk_ingestion)
    task_projects.set_downstream(task_disable_bulk_ingestion)

    conf_namespaces = get_namespaces(variable_name="namespaces_bulk_ingestion")
    conf_projects = get_projects(variable_name="projects_bulk_ingestion")

    task_export = PythonOperator(
        task_id="articles",
        python_callable=articles(
            projects=conf_projects,
            namespaces=conf_namespaces,
        ),
        execution_timeout=timedelta(days=10),
    )

    task_export.set_upstream(task_verify_bulk_ingestion)
    task_export.set_downstream(task_disable_bulk_ingestion)

    task_enable_article_bulk = PythonOperator(
        task_id="enable_article_bulk",
        python_callable=enable_article_bulk_service,
        execution_timeout=timedelta(seconds=30),
    )

    task_enable_article_bulk.set_upstream(task_disable_bulk_ingestion)

    task_monitor_article_bulk = PythonOperator(
        task_id="monitor_article_bulk",
        python_callable=monitor_article_bulk_service,
        execution_timeout=timedelta(days=30),
    )

    task_monitor_article_bulk.set_upstream(task_enable_article_bulk)

    task_disable_article_bulk = PythonOperator(
        task_id="disable_article_bulk",
        python_callable=disable_article_bulk_service,
        execution_timeout=timedelta(seconds=30),
    )

    task_disable_article_bulk.set_upstream(task_monitor_article_bulk)
