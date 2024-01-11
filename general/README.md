# Wikimedia Enterprise General

This directory houses the packages used across the project's APIs and services:

1. [config](/general/config/) - provides the general configuration for the project (languages, etc.)

2. [httputil](/general/httputil/) - exposes custom middleware shared across APIs

3. [ksqldb](/general/ksqldb/) - serves as a client for connecting to the `kslqDB` database

4. [log](/general/log/) - acts as a log wrapper to abstract the underlying logger and make it replaceable

5. [parser](/general/parser/) - contains the `Parsoid HTML` parser that extracts data from content

6. [prometheus](/general/prometheus/) - wraps the `Prometheus` library to streamline the exposure of metrics

7. [protos](/general/protos/) - houses `.proto` files that describe service interfaces for services communicating through `gRPC`

8. [schema](/general/schema/) - Contains the `Avro` schema for the services and schema registry communication helpers

9. [subscriber](/general/subscriber/) - streamlines `Kafka` connection for the services

10. [wmf](/general/wmf/) - wraps to streamline communication with `WMF` APIs
