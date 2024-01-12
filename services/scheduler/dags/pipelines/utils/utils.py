from os import getenv, path
from json import load
from logging import getLogger
from random import choice
from time import sleep
from dns.resolver import query
from airflow.models import Variable
from grpc import insecure_channel
import boto3


def get_channel(addr_name):
    """
    Returns a secure channel to the grpc server
    if root certificates, private key and certificate chain
    are found in the env variables,
    otherwise returns an insecure channel.
    """
    logger = getLogger("airflow.task")

    # Server target and certificates for mutual TLS.
    addr = getenv(addr_name)
    # rt_cert = getenv("MTLS_CLIENT_CERT_PEM")
    # pr_key = getenv("MTLS_CLIENT_PRIVATE_KEY_PEM")
    # cert_chain = getenv("INTERNAL_ROOT_CA_PEM")

    dns_addr, port = get_addr_and_port(addr=addr)
    ips = []  # this will hold service A records

    for server in query(dns_addr, "A"):
        ips.append(str(server))

    ip_addr = "{ip}:{port}".format(
        # choosing random record to simulate load balancer
        ip=choice(ips), port=port
    )

    logger.info(
        "Trying to get channel with env variable `{addr_name}` with address `{addr}` on ip address `{ip_addr}`\n".format(
            addr_name=addr_name, addr=addr, ip_addr=ip_addr
        )
    )

    return insecure_channel(ip_addr)


__languages = {}


def get_language(project):
    """
    Makes a lookup into into languages dictionary to find a project language.

    Returns:
        string: Language code
    """

    if len(__languages) == 0:
        with open(
            path.join(
                path.dirname(__file__),
                "../../../general/config/languages.json",
            ),
            "r",
        ) as file:
            languages = load(file)

            for identifier in languages:
                __languages[identifier] = languages[identifier]

    return __languages[project]


def get_namespaces(variable_name="namespaces"):
    """
    Returns a list of namespaces from Airflow variable 'namespaces'.
    If the `namespaces` variable not set, it uses config/namespaces.json file.

    Returns:
        list: List of namespaces.
    """
    logger = getLogger("airflow.task")

    try:
        namespaces = Variable.get(variable_name, deserialize_json=True)
    except KeyError:
        with open(
            path.join(
                path.dirname(__file__),
                "../../../general/config/namespaces.json",
            ),
            "r",
        ) as file:
            namespaces = load(file)

    logger.info(
        "Got namespaces configuration: `{namespaces}`".format(
            namespaces=namespaces)
    )

    return namespaces


def get_projects(variable_name="projects"):
    """
    Returns a list of projects by parsing projects json from
    Airflow variable 'projects'.
    If 'projects' variable is not set, it uses config/random_projects.json file.

    Returns:
        list: List of projects.
    """
    logger = getLogger("airflow.task")
    projects = []

    try:
        projects = Variable.get(variable_name, deserialize_json=True)
    except KeyError:
        with open(
            path.join(
                path.dirname(__file__),
                "../../../general/config/random_projects.json",
            ),
            "r",
        ) as file:
            projects = load(file)

    logger.info("Got projects configuration: `{projects}`".format(
        projects=projects))

    return projects


def get_exclude_events():
    """Returns a list of event types to exclude"""

    exclude_events = []

    try:
        exclude_events = Variable.get("exclude_events", deserialize_json=True)
    except KeyError:
        return ["delete"]

    return exclude_events


def get_batch_size(variable_name="batch_size"):
    """Returns a batch size variable to control concurrent ingestion of projects"""

    batch_size = None

    try:
        batch_size = Variable.get(variable_name)
    except KeyError:
        return 150

    return int(batch_size)


def get_copy_workers():
    """Returns numbers of workers to control concurrency for copy job"""

    __max_workers = None

    try:
        __max_workers = int(Variable.get("copy_max_workers"))
    except (ValueError, KeyError):
        __max_workers = 25

    return __max_workers


def run_copy():
    """Returns boolean run_copy. For testing purporses."""

    __run_copy = None

    try:
        __run_copy = Variable.get("run_copy")

        if __run_copy == "true":
            __run_copy = True
        else:
            __run_copy = False

    except (ValueError, KeyError):
        __run_copy = False

    return __run_copy


def chunks(lst, n):
    """Return successive n-sized chunks from list."""
    return [lst[i: i + n] for i in range(0, len(lst), n)]


def get_addr_and_port(addr):
    """Splits service discovery domain into port and address"""
    parts = addr.split(":")
    return parts[0], parts[1]


def get_ip_addresses(addr_name, return_as_iterator=True):
    """
    Given an address string, returns an iterator that returns an incremental
    element of the list of IP addresses associated with the address.

    Args:
        addr_name (str): The address to look up IP addresses for.

    Returns:
        iterator: An iterator that returns an incremental element of the list of IP addresses.
    """
    # Get the DNS address from environment variable
    addr = getenv(addr_name)

    # Extract DNS address and port from input address
    dns_addr, port = get_addr_and_port(addr=addr)

    # Query the DNS server for IP addresses associated with the DNS address
    ips = []

    for server in query(dns_addr, "A"):
        ips.append("{server}:{port}".format(server=str(server), port=port))

    # Sorting the list to get consistent results
    ips.sort()

    # Return an iterator that returns an incremental element of the list of IP addresses
    return get_list_iterator(ips) if return_as_iterator else ips


def get_list_iterator(lst):
    """
    Returns an iterator that iterates over the given list indefinitely, returning each element one at a time
    in an incremental manner (i.e. the first call returns the first element, the second call returns the second
    element, and so on). When the end of the list is reached, the iterator starts back at the beginning of the list.

    Args:
        lst: The list to iterate over.

    Returns:
        An iterator that iterates over the list indefinitely, returning each element one at a time in an incremental
        manner.

    Example Usage:
        >>> my_list = [1, 2, 3, 4]
        >>> my_iter = get_list_iterator(my_list)
        >>> next(my_iter)
        1
        >>> next(my_iter)
        2
        >>> next(my_iter)
        3
        >>> next(my_iter)
        4
        >>> next(my_iter)
        1
        >>> next(my_iter)
        2
        >>> next(my_iter)
        3
        >>> next(my_iter)
        4
        >>> # and so on...
    """
    index = 0

    while True:
        yield lst[index]  # yield the current element
        index += 1  # increment the index
        if index >= len(
            lst
        ):  # if we've reached the end of the list, start over at the beginning
            index = 0


def ecs_service_set_task_count(ecs_cluster_name, ecs_service_name, desired_count):
    """
    Uses boto3 library to set an amount of running instances per ECS service.
    Initialy is used for services taht are running batch jobs, and do not need to run 24/7.
    Boto3 library takes care about getting the access to AWS API - via environment variables with
    credentials or task role policy with ecs:UpdateService permissions.
    """
    result = False

    logger = getLogger("airflow.task")

    try:
        # create ecs client
        ecs_client = boto3.client("ecs")

        # update the service config
        ecs_client.update_service(
            cluster=ecs_cluster_name,
            service=ecs_service_name,
            desiredCount=desired_count,
        )

        # log succeful event
        logger.info(
            "Successfuly updated ECS service configuration: "
            f"{ecs_cluster_name}:{ecs_service_name}, desired_count: {desired_count}"
        )

        result = True

    except Exception as exc:
        # log an exception
        logger.error(f"Failed to update the ECS service configuration: {exc}")

    return result


def get_secret_value(secret_id):
    """
    Gets the secret value from AWS Secrets Manager.

    Returns:
        string: Secret value
    """
    secret_client = boto3.client("secretsmanager")

    response = secret_client.get_secret_value(SecretId=secret_id)

    if 'SecretString' in response:
        return response['SecretString']

    return ""


def verify_ip_addresses_returned_by_dns(dns_name, desired_count, attempts=30, delay=15):
    """
    Verify if internal service discovery is able to return :desired_count of IP addresses,
    for :attempts times, and with delay of :delay.
    """

    class Result:
        def __init__(self):
            self.success = False
            self.ips = []

        def set_successful_response(self):
            self.success = True
            return self

        def set_ips(self, ips):
            self.ips = ips
            return self

    result = Result()

    logger = getLogger("airflow.task")

    logger.info(
        f"Verification: waiting for dns name {dns_name} to return {desired_count} IPs"
    )

    for attempt in range(1, attempts + 1):
        resolved_ips = []

        try:
            resolved_ips = get_ip_addresses(dns_name, return_as_iterator=False)
            logger.info(f"Attempt {attempt}: DNS query: result {resolved_ips}")
        except Exception as exc:
            logger.warning(
                f"Attempt {attempt}: DNS query: exception was caught execution: {exc}"
            )

        if len(resolved_ips) == desired_count:
            # verification succeeded
            logger.info(f"dns name {dns_name} returned ips: {resolved_ips}")
            result.set_successful_response().set_ips(resolved_ips)
            break
        else:
            if attempt == attempts:
                # verification has failed
                logger.error(
                    f"Verification has failed after {attempts} attempts")
                break
            else:
                # wait a bit more
                logger.info(f"Waiting for {delay} seconds")
                sleep(delay)

    return result


def get_desired_count(
        env_variable_name="ECS_SNAPSHOT_SERVICE_DESIRED_COUNT",
        airflow_variable_name="desired_count_snapshots",
        default_desired_count=3):
    """
    Retrieves the desired count for a service from either an Airflow Variable or an environment variable.
    If both sources fail to provide a value, returns a default desired count.

    Args:
        env_variable_name: The name of the environment variable to retrieve the desired count from.
                            Defaults to "ECS_SNAPSHOT_SERVICE_DESIRED_COUNT".
        airflow_variable_name: The name of the Airflow Variable to retrieve the desired count from.
                               Defaults to "desired_count_snapshots".
        default_desired_count: The default desired count to return if neither source provides a value.
                               Defaults to 3.

    Returns:
        The desired count for the service.
    """
    logger = getLogger("airflow.task")
    desired_count = None

    try:
        desired_count = int(Variable.get(airflow_variable_name))
    except Exception:
        desired_count = None
        logger.info(
            f"failed to retrieve desired count from airflow {airflow_variable_name} ")

    if desired_count is None:
        try:
            desired_count = int(getenv(env_variable_name))
        except Exception:
            desired_count = None
            logger.info(
                f"failed to retrieve desired count from env {env_variable_name} ")

    logger.info(
        f"retrieved desired count {desired_count}, var: {airflow_variable_name}, env: {env_variable_name}")

    return default_desired_count if desired_count is None else desired_count
