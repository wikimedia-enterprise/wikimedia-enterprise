# Wikimedia Enterprise General

This directory houses the packages used across the project's APIs and services:

1. [/config](/general/config/) - provides the general configuration for the project (languages, etc.)

1. [/httputil](/general/httputil/) - exposes custom middleware shared across APIs

1. [/ksqldb](/general/ksqldb/) - serves as a client for connecting to the `kslqDB` database

1. [/log](/general/log/) - acts as a log wrapper to abstract the underlying logger and make it replaceable

1. [/parser](/general/parser/) - contains the `Parsoid HTML` parser that extracts data from content

1. [/prometheus](/general/prometheus/) - wraps the `Prometheus` library to streamline the exposure of metrics

1. [/protos](/general/protos/) - houses `.proto` files that describe service interfaces for services communicating through `gRPC`

1. [/schema](/general/schema/) - contains the `Avro` schema for the services and schema registry communication helpers

1. [/subscriber](/general/subscriber/) - streamlines `Kafka` connection for the services

1. [/wmf](/general/wmf/) - wraps to streamline communication with `WMF` APIs
