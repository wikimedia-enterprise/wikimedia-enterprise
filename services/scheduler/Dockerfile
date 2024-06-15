FROM apache/airflow:slim-2.8.3-python3.10

# Default AirFlow dir
ARG AIRFLOW_DIR=/opt/airflow

# Set AirFlow user and group id
ENV AIRFLOW_UID=50000 AIRFLOW_GID=0

# Switch to /airflow dir
WORKDIR /airflow

# Copy and install requirements into /airflow
COPY ./services/scheduler/requirements.txt .
RUN pip install -r requirements.txt

# Copy scripts into /airflow
COPY ./services/scheduler/scripts/ .

# Copy dags and plugins directories content into airflow dir
COPY --chown=airflow:root ./services/scheduler/dags ${AIRFLOW_DIR}/dags
COPY --chown=airflow:root ./services/scheduler/plugins ${AIRFLOW_DIR}/plugins
COPY --chown=airflow:root ./general/config ${AIRFLOW_DIR}/general/config

# Switch into airflow user and start webserver and scheduler
USER airflow
CMD [ "bash", "./start.sh" ]
