.PHONY: airflow protos

airflow:
	docker-compose up -d airflow

protos:
	python3 -m grpc_tools.protoc \
	--proto_path=pipelines/protos=submodules/protos \
	--python_out=dags/ \
	--grpc_python_out=dags/ \
	submodules/protos/bulk.proto \
	submodules/protos/hourlys.proto \
	submodules/protos/snapshots.proto
