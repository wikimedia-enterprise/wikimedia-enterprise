#!/usr/bin/env bash

echo "[INFO] running db init"
airflow db init

echo "[INFO] starting scheduler in background"
airflow scheduler &

echo "[INFO] starting webserver"
exec airflow webserver
