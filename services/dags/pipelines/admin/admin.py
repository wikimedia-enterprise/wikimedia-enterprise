"""
DAG for admin service. Contains Delete task, to be invoked manually.
The articles to be used come from the variables "delete_articles" and "delete_articles_s3url". Only one may be set on a given invokation.
"""

import sys, csv, json, io
from os import getenv, path
from datetime import timedelta, datetime
from google.protobuf.json_format import MessageToDict
from airflow import DAG, AirflowException
from airflow.operators.python import PythonOperator
from airflow.providers.amazon.aws.hooks.s3 import S3Hook
from airflow.models import Variable
from grpc import insecure_channel
from airflow.settings import AIRFLOW_HOME

DAGS_HOME = path.join(AIRFLOW_HOME, "dags/git_wme-dags")
WORKERS_HOME = path.join(AIRFLOW_HOME, "dags/dags.git")

sys.path += [DAGS_HOME, WORKERS_HOME]
print(f"sys.path: {sys.path}")

from logging import getLogger

from pipelines.protos.admin_pb2_grpc import AdminStub
from pipelines.protos.admin_pb2 import (
    ArticleIdentifier,
    DeleteRequest,
)

admin_service_address = "0.0.0.0:5051"
environment  = getenv('CLUSTER_ENV')
pod_template_file = path.join(DAGS_HOME, f'pod_templates/admin-{environment}.yaml')

# Initialize logger
logger = getLogger("airflow.task")

def make_admin_grpc_request():
    """
    Calls Admin.Delete based on the value of the delete_articles and delete_articles_s3url variables.
    """
    req = get_articles_to_delete()
    with insecure_channel(admin_service_address) as channel:
        logger.info("Making `Admin.Delete` request.")

        stub = AdminStub(channel)
        res = stub.Delete(req, wait_for_ready=True)

        logger.info(f"Finished `Admin.Delete` request. Response:\n{res}\n")

def article_identifier_from_csv_row(row):
    if len(row) != 3:
        raise ValueError(
            "CSV must have exactly three fields: database, namespace, title"
        )
    article = ArticleIdentifier()
    article.database = row[0].strip()
    article.namespace = int(row[1].strip())
    article.title = row[2].strip()
    return article

def proto_to_str(msg: ArticleIdentifier) -> str:
    """Returns a nicely formatted message for printing."""
    dictionary = MessageToDict(msg)
    # including_default_value_fields is not available in our version. namespace is often 0 (default),
    # so let's make an extra effort to print it.
    dictionary["namespace"] = msg.namespace
    return json.dumps(dictionary)

def article_identifier_from_tuple(tuple: tuple[str, int, str]) -> ArticleIdentifier:
    return ArticleIdentifier(
        database=tuple[0],
        namespace=tuple[1],
        title=tuple[2],
    )

def read_from_s3(articles_s3url):
    if not articles_s3url.startswith("s3://"):
        raise ValueError(
            f"S3 URL must start with s3://. Value: {articles_s3url}"
        )
    path = articles_s3url[len("s3://"):]
    parts = path.split("/")
    bucket = parts[0]
    file = "/".join(parts[1:])

    logger.info("Retrieving article list")
    logger.info("Bucket: " + bucket)
    logger.info("File: " + file)

    s3 = S3Hook()
    obj = s3.get_key(file, bucket)
    content = obj.get()['Body'].read().decode('utf-8')
    return content

def get_articles_to_delete(
    articles_var="delete_articles", articles_s3url_var="delete_articles_s3url"
) -> DeleteRequest:
    """
    Processes the relevant Scheduler variables and, if necessary, reads the relevant CSV file to retrieve the article list to delete.

    Fails unless exactly one of the variables is set.
    Args:
        articles_var (str): Name of the Airflow variable containing a list of articles to delete. Format: [["enwiki", 0, "Josephine_Baker"], ...]
        articles_s3url_var (str): Name of the Airflow variable containing the S3 URL to a CSV file, containing the list of articles to delete.
          For example: s3://wme-eks-admin-dv/path/to/file.csv
          admin-service should only use s3 buckets wme-eks-admin-dv and wme-eks-admin-pr.
          The file must have three columns: database, namespace, title. For example: "enwiki, 0, Josephine_Baker".
    Returns:
        The gRPC request to send to AdminService.
    """
    articles = Variable.get(articles_var, "")
    articles_s3url = Variable.get(articles_s3url_var, "")

    if bool(articles) == bool(articles_s3url):
        raise AirflowException(
            f"Exactly one of {articles_var} and {articles_s3url_var} must be set.\n\n"
            f"Values:\n"
            f"{articles_var}={articles}\n"
            f"{articles_s3url_var}={articles_s3url}"
        )

    req = DeleteRequest()
    if articles:
        try:
            parsed_articles = json.loads(articles)
        except Exception as e:
            raise ValueError("Failed to decode '%s'" % (articles)) from e
        req.articles.extend([article_identifier_from_tuple(x) for x in parsed_articles])

        art_str = proto_to_str(req.articles[0])
        logger.info(f"Running take-down pipeline for articles [{art_str}, ...] (size={len(req.articles)})")
    else:
        contents = read_from_s3(articles_s3url)

        req = DeleteRequest()
        csvfile = io.StringIO(contents)

        reader = csv.reader(csvfile)
        for row in reader:
            req.articles.append(article_identifier_from_csv_row(row))

        art1_str = proto_to_str(req.articles[0])
        art2_str = proto_to_str(req.articles[-1])
        logger.info(
            f"Running take-down pipeline for articles [{art1_str}, ..., {art2_str}] (size={len(req.articles)})"
        )

    return req


with DAG(
    dag_id="admin",
    start_date=datetime(2022, 1, 1, 0, 0, 0),
    schedule_interval=None,
    default_args=dict(
        retries=1,
        retry_delay=timedelta(seconds=30),
    )
) as dag:
    task_make_grpc_request = PythonOperator(
        task_id="make_admin_grpc_request",
        python_callable=make_admin_grpc_request,
        executor_config = {
            "pod_template_file": pod_template_file
        }
    )
