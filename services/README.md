# Wikimedia Enterprise Services

This directory houses the services used by the project:

1. [bulk-ingestion](/services/bulk-ingestion/) - feeds baseline data into the system

2. [content-integrity](/services/content-integrity/) - manages the content integrity domain

3. [event-bridge](/services/event-bridge/) - connects to the WMF event stream

4. [on-demand](/services/on-demand/) - stores data for the `Ondemand API` in the object store

5. [scheduler](/services/scheduler/) - schedules ingestion and processing pipelines

6. [snapshots](/services/snapshots/) - creates snapshots and stores them in the object store

7. [structured-data](/services/structured-data/) - calls WMF APIs, enriches the data, and propagates it throughout the system
