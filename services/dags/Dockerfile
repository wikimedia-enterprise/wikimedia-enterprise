FROM apache/airflow:slim-2.9.3-python3.10

# Default AirFlow dir
ARG AIRFLOW_DIR=/opt/bitnami/airflow

# Set AirFlow user and group id
ENV AIRFLOW_UID=50000 AIRFLOW_GID=0

# Switch to /airflow dir
WORKDIR /airflow

# Copy and install requirements into /airflow
COPY requirements.txt .
RUN pip install -r requirements.txt

# Copy dags and plugins directories content into airflow dir
COPY --chown=airflow:root submodules/config ${AIRFLOW_DIR}/submodules/config

# Switch into airflow user and start webserver and scheduler
USER airflow
