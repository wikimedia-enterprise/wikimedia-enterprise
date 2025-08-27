# Wikimedia Enterprise Protos

This repository contains `proto` files that will be used for generating server and client code across projects.

### Steps to update Python stubs for Protobuf generated files:

1. In terminal, move to the `protos` project folder from you project root folder: `cd general/protos`
2. Create a python virtual environment to install protobuf dependencies: `python3 -m venv venv`
3. Activate virtual environment: `. venv/bin/activate`
4. Install protobuf dependencies: `pip install grpcio grpcio-tools`
5. Generate the python stub for the snapshot proto:
```
python -m grpc_tools.protoc \
  --proto_path=. \
  --python_out=. \
  --grpc_python_out=. \
  snapshots.proto
```
  If you want to generate all the stubs, you can use the following command:
```
python -m grpc_tools.protoc  \
   --proto_path=. \
   --python_out=.\
   --grpc_python_out=.\
  *.proto
```

6. Deactivate virtual environment: `deactivate`
7. Copy and overwrite the old pb2 files in `protos` folder with the new ones.
8. Remove the generated python stubs from the protos project folder: `rm -rf __pycache__ *.pyc`
