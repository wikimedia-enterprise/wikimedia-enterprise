"""
DAG for cognito user management. Disables and deletes unconfirmed cognito users.
Disables and deletes users with usernames that match a regex list, if provided.
Disables and deletes users with emails that match a domain regex list, if provided.
"""

from os import getenv
from datetime import datetime, timedelta, timezone
import json
from logging import getLogger
import boto3
from botocore.exceptions import ClientError
from airflow import DAG
from airflow.models import Variable
from airflow.operators.python_operator import PythonOperator
from pipelines.utils.utils import (
    check_domain,
    match_patterns,
)

# Initialize logger
logger = getLogger("airflow.task")


def remove_users(user_pool_id, **kwargs):
    """disable and delete users from a user pool

    Args:
        user_pool_id (string): user pool id
    """
    client = boto3.client("cognito-idp")

    # List of user groups to manage
    groups = Variable.get("user_groups", default_var=["group_1"], deserialize_json=True)

    # Confirmation window in days
    confirmation_window = int(Variable.get("user_confirmation_window", default_var=1))

    # List of email domain regexs
    domains = Variable.get("user_domains", default_var=[], deserialize_json=True)

    # List of username regexs
    usernames = Variable.get("users", default_var=[], deserialize_json=True)

    # To allow for some time for legitimate users to finish ongoing signup
    window = datetime.now(timezone.utc) - timedelta(days=confirmation_window)

    for group in groups:
        pagination_token = None

        while True:
            try:
                if pagination_token:
                    response = client.list_users_in_group(
                        UserPoolId=user_pool_id,
                        GroupName=group,
                        NextToken=pagination_token,
                    )
                else:
                    response = client.list_users_in_group(
                        UserPoolId=user_pool_id, GroupName=group
                    )
            except ClientError as error:
                logger.error(
                    f"List user API call failed for group {group} with error: {json.dumps(error.response)}"
                )
                continue

            for user in response["Users"]:
                # default behaviour: always remove unconfirmed users
                if user["UserStatus"] == "UNCONFIRMED":
                    if user["UserCreateDate"] <= window:
                        disable_and_delete_user(client, user_pool_id, user["Username"])

                # if domain regexes set, remove those users
                if len(domains):
                    for attribute in user["Attributes"]:
                        if attribute["Name"] == "email" and check_domain(
                            domains, attribute["Value"]
                        ):
                            disable_and_delete_user(
                                client, user_pool_id, user["Username"]
                            )

                # if usernames regexes set, remove those users
                if len(usernames):
                    if match_patterns(usernames, user["Username"]):
                        disable_and_delete_user(client, user_pool_id, user["Username"])

            pagination_token = response.get("NextToken")

            if not pagination_token:
                break


def disable_and_delete_user(client, user_pool_id, username):
    """disable and delete user from a user pool

    Args:
        client (object): cognito client
        user_pool_id (string): user pool id
        username (string): username
    """
    try:
        client.admin_disable_user(UserPoolId=user_pool_id, Username=username)
    except ClientError as error:
        logger.error(
            f"Disable user API call failed for user {username} with error: {json.dumps(error.response)}"
        )

    try:
        client.admin_delete_user(UserPoolId=user_pool_id, Username=username)
    except ClientError as error:
        logger.error(
            f"Delete user API call failed for user {username} with error: {json.dumps(error.response)}"
        )

    logger.info(f"Successfully disabled and deleted user {username}")


dag_args = dict(
    dag_id="user_management",
    start_date=datetime(2024, 1, 1, 0, 0, 0),
    schedule_interval=timedelta(days=90),
    catchup=False,
    default_args=dict(provide_context=True),
)

with DAG(**dag_args) as dag:
    task_user_cleanup = PythonOperator(
        task_id="user_cleanup",
        python_callable=remove_users,
        op_kwargs=dict(user_pool_id=getenv("COGNITO_USER_POOL_ID")),
        execution_timeout=timedelta(hours=2),
    )
