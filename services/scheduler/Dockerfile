FROM apache/airflow:2.2.5

# Default AirFlow dir
ARG AIRFLOW_DIR=/opt/airflow

# Set AirFlow user and group id
ENV AIRFLOW_UID=50000 AIRFLOW_GID=0

# Switch to /airflow dir
WORKDIR /airflow

# Copy and install requirements into /airflow
COPY requirements.txt .
RUN pip install -r requirements.txt

# Copy scripts into /airflow
COPY scripts/ .

# Copy dags and plugins directories content into airflow dir
COPY --chown=airflow:root dags ${AIRFLOW_DIR}/dags
COPY --chown=airflow:root plugins ${AIRFLOW_DIR}/plugins
COPY --chown=airflow:root submodules/config ${AIRFLOW_DIR}/submodules/config

# Switch into airflow user and start webserver and scheduler
USER airflow
CMD [ "bash", "./start.sh" ]
