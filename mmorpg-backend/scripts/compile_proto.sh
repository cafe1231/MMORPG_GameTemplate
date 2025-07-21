#!/bin/bash

# Protocol Buffer Compilation Script
# Compiles .proto files for both Go backend and C++ Unreal Engine client

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Paths
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
PROTO_DIR="$PROJECT_ROOT/pkg/proto"
GO_OUT_DIR="$PROJECT_ROOT/pkg/proto"
CPP_OUT_DIR="$PROJECT_ROOT/../UnrealEngine/Plugins/MMORPGTemplate/Source/MMORPGCore/Public/Proto"

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
    echo -e "${RED}Error: protoc is not installed${NC}"
    echo "Please install Protocol Buffers compiler:"
    echo "  - macOS: brew install protobuf"
    echo "  - Ubuntu: sudo apt-get install protobuf-compiler"
    echo "  - Windows: Download from https://github.com/protocolbuffers/protobuf/releases"
    exit 1
fi

# Check if Go protoc plugin is installed
if ! command -v protoc-gen-go &> /dev/null; then
    echo -e "${YELLOW}Installing Go protoc plugin...${NC}"
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi

# Check if Go gRPC plugin is installed
if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo -e "${YELLOW}Installing Go gRPC plugin...${NC}"
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
fi

echo -e "${GREEN}Starting Protocol Buffer compilation...${NC}"

# Create output directories if they don't exist
mkdir -p "$GO_OUT_DIR"
mkdir -p "$CPP_OUT_DIR"

# Compile for Go
echo -e "${YELLOW}Compiling for Go...${NC}"
cd "$PROJECT_ROOT"

for proto_file in "$PROTO_DIR"/*.proto; do
    if [ -f "$proto_file" ]; then
        echo "  Compiling $(basename "$proto_file")..."
        protoc \
            --go_out="$GO_OUT_DIR" \
            --go_opt=paths=source_relative \
            --go-grpc_out="$GO_OUT_DIR" \
            --go-grpc_opt=paths=source_relative \
            -I "$PROTO_DIR" \
            "$proto_file"
    fi
done

echo -e "${GREEN}Go compilation complete!${NC}"

# Compile for C++ (Unreal Engine)
echo -e "${YELLOW}Compiling for C++ (Unreal Engine)...${NC}"

# Check if C++ output directory exists
if [ ! -d "$CPP_OUT_DIR" ]; then
    echo -e "${YELLOW}Warning: C++ output directory doesn't exist. Creating it...${NC}"
    mkdir -p "$CPP_OUT_DIR"
fi

for proto_file in "$PROTO_DIR"/*.proto; do
    if [ -f "$proto_file" ]; then
        echo "  Compiling $(basename "$proto_file")..."
        protoc \
            --cpp_out="$CPP_OUT_DIR" \
            -I "$PROTO_DIR" \
            "$proto_file"
    fi
done

echo -e "${GREEN}C++ compilation complete!${NC}"

# Generate a summary header for C++
echo -e "${YELLOW}Generating C++ include header...${NC}"
cat > "$CPP_OUT_DIR/MMORPGProto.h" << EOF
// Auto-generated header file
// Include all Protocol Buffer headers

#pragma once

#include "base.pb.h"
#include "auth.pb.h"
#include "character.pb.h"
#include "world.pb.h"
#include "game.pb.h"
#include "chat.pb.h"

namespace MMORPG
{
    // Type aliases for easier use in Unreal Engine
    using namespace mmorpg;
}
EOF

echo -e "${GREEN}Protocol Buffer compilation completed successfully!${NC}"
echo ""
echo "Generated files:"
echo "  - Go files: $GO_OUT_DIR"
echo "  - C++ files: $CPP_OUT_DIR"
echo ""
echo "To use in your code:"
echo "  - Go: import \"github.com/mmorpg-template/backend/pkg/proto\""
echo "  - C++: #include \"Proto/MMORPGProto.h\""