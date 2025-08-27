.PHONY: airflow build protos lint

protos:
	python3 -m grpc_tools.protoc \
	--proto_path=pipelines/protos=submodules/protos \
	--python_out=./ \
	--grpc_python_out=./ \
	submodules/protos/bulk.proto \
	submodules/protos/snapshots.proto \
	submodules/protos/admin.proto

lint:
	flake8 --exclude dags/pipelines/protos/ --max-line-length 160 --extend-ignore=W293,W291,W391,E203
