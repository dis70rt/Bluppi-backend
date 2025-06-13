.PHONY: proto

	FILE

proto:
	@echo "Generating Python code for $(FILE)..."
	python -m grpc_tools.protoc \
		--python_out=. \
		--grpc_python_out=. \
		--proto_path=. \
		$(FILE)
	@echo "Generated files for $(FILE)"

install-deps:
	@echo "Installing protobuf dependencies..."
	pip install grpcio grpcio-tools protobuf
	@echo "Dependencies installed"