#!/bin/bash

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

VENV_DIR=".venv"

source "$VENV_DIR/bin/activate"    

generate_proto() {
    local FILE=$1
    echo -e "${YELLOW}Generating Python code for $FILE...${NC}"
    
    python -m grpc_tools.protoc \
        --python_out=. \
        --grpc_python_out=. \
        --pyi_out=. \
        --proto_path=. \
        "$FILE"
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}Generated files for $FILE${NC}"
        return 0
    else
        echo -e "${RED}Failed to generate files for $FILE${NC}"
        return 1
    fi
}

cd party

if ! python -c "import grpc_tools" 2>/dev/null; then
    echo -e "${YELLOW}Installing protobuf dependencies...${NC}"
    pip install grpcio grpcio-tools protobuf mypy-protobuf
    echo -e "${GREEN}Dependencies installed${NC}"
fi

PROTO_FILES=$(find protobuf -name "*.proto" | sort)

if [ -z "$PROTO_FILES" ]; then
    echo -e "${RED}No .proto files found in protobuf directory${NC}"
    exit 1
fi

echo -e "${GREEN}Found proto files:${NC}"
for FILE in $PROTO_FILES; do
    echo "  $FILE"
done

FAILED_FILES=()
for FILE in $PROTO_FILES; do
    if ! generate_proto "$FILE"; then
        FAILED_FILES+=("$FILE")
    fi
done

if [ ${#FAILED_FILES[@]} -eq 0 ]; then
    echo -e "${GREEN}✓ All files generated successfully!${NC}"
    echo -e "${GREEN}Generated file types:${NC}"
    
    for FILE in $PROTO_FILES; do
        BASE_NAME=$(basename "$FILE" .proto)
        DIR_NAME=$(dirname "$FILE")
        echo "  ${DIR_NAME}/${BASE_NAME}_pb2.py (messages)"
        echo "  ${DIR_NAME}/${BASE_NAME}_pb2.pyi (type stubs)"
        if grep -q "service " "$FILE"; then
            echo "  ${DIR_NAME}/${BASE_NAME}_pb2_grpc.py (services)"
        fi
    done
else
    echo -e "${RED}✗ Failed to generate: ${FAILED_FILES[*]}${NC}"
    exit 1
fi

cd ..
echo -e "${GREEN}All done! 🎉${NC}"
echo -e "${YELLOW}Note: .pyi files provide better IDE support and type checking${NC}"