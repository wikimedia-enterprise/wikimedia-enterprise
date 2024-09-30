.PHONY: airflow protos lint

airflow:
	docker compose up -d airflow

protos:
	python3 -m grpc_tools.protoc \
	--proto_path=pipelines/protos=submodules/protos \
	--python_out=dags/ \
	--grpc_python_out=dags/ \
	submodules/protos/bulk.proto \
	submodules/protos/snapshots.proto

lint:
	flake8 --exclude dags/pipelines/protos/ --max-line-length 160 --extend-ignore=W293,W291,W391,E203

