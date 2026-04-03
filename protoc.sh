#!/bin/bash

PROTO_DIR="internals/proto"
GEN_DIR="internals/gen"

GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

SPINNER=("⠋" "⠙" "⠹" "⠸" "⠼" "⠴" "⠦" "⠧" "⠇" "⠏")

execute_command() {
    local label="$1"
    shift
    local cmd=("$@")
    
    local err_file=$(mktemp)
    "${cmd[@]}" 2> "$err_file" &
    local pid=$!
    
    local i=0
    while kill -0 "$pid" 2>/dev/null; do
        printf "\r${BLUE}${SPINNER[i]}${NC} ${label}..."
        i=$(( (i + 1) % 10 ))
        sleep 0.1
    done

    wait "$pid"
    local exit_code=$?
    
    if [ $exit_code -eq 0 ]; then
        printf "\r${GREEN}✔${NC} ${label}     \n"
    else
        printf "\r${RED}✖${NC} ${label}     \n"
        printf "${RED}Error details:${NC}\n"
        cat "$err_file"
        rm "$err_file"
        return 1
    fi
    rm "$err_file"
}

if [ ! -d "$PROTO_DIR" ]; then
    printf "${RED}Error: Proto directory '$PROTO_DIR' not found.${NC}\n"
    exit 1
fi

rm -rf "$GEN_DIR"
mkdir -p "$GEN_DIR"

echo -e "${YELLOW}Starting Protocol Buffers generation...${NC}\n"

found_files=false
for proto_file in "$PROTO_DIR"/*.proto; do
    if [ -f "$proto_file" ]; then
        found_files=true
        filename=$(basename "$proto_file")
        
        cmd=(
            protoc 
            --proto_path="$PROTO_DIR"
            --go_out=. 
            --go-grpc_out=. 
            "$filename"
        )
        
        execute_command "Generating $filename" "${cmd[@]}"
        
        if [ $? -ne 0 ]; then
            exit 1
        fi
    fi
done

if [ "$found_files" = false ]; then
    printf "${YELLOW}No .proto files found in $PROTO_DIR${NC}\n"
else
    echo -e "\n${GREEN}✨ All proto files generated successfully! ✨${NC}"
fi