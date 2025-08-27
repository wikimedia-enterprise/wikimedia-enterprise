#!/bin/bash

# Configuration
BASE_URL="git@gitlab.enterprise.wikimedia.com:wikimedia-enterprise"

api_projects=(
    auth
    main
    realtime
)
general_projects=(
    config
    httputil
    ksqldb
    log
    parser
    prometheus
    protos
    schema
    subscriber
    tracing
    wmf
)
services_projects=(
    bulk-ingestion
    content-integrity
    dags
    eventstream-listener
    on-demand
    snapshots
    structured-data
)

for project in "${api_projects[@]}"; do
    cd api

    echo "processing $project"
    rm -rf $project
    git clone "$BASE_URL/api/$project" $project

    cd -
done


for project in "${general_projects[@]}"; do
    cd general

    echo "processing $project"
    rm -rf $project
    git clone "$BASE_URL/general/$project" $project

    cd -
done


for project in "${services_projects[@]}"; do
    cd services

    echo "processing $project"
    rm -rf $project
    git clone "$BASE_URL/services/$project" $project

    cd -
done
