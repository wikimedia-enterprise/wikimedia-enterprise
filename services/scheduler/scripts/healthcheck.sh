#!/usr/bin/env bash

curl --fail http://localhost:8080/health
airflow jobs check --job-type SchedulerJob --hostname "$${HOSTNAME}"
