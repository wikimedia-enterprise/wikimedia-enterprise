# Wikimedia Enterprise Services

This directory houses the services used by the project:

1. [/bulk-ingestion](/services/bulk-ingestion/) - feeds baseline data into the system

1. [/content-integrity](/services/content-integrity/) - manages the content integrity domain

1. [/event-bridge](/services/event-bridge/) - connects to the WMF event stream

1. [/on-demand](/services/on-demand/) - stores data for the `Ondemand API` in the object store

1. [/scheduler](/services/scheduler/) - schedules ingestion and processing pipelines

1. [/snapshots](/services/snapshots/) - creates snapshots and stores them in the object store

1. [/structured-data](/services/structured-data/) - calls WMF APIs, enriches the data, and propagates it throughout the system
