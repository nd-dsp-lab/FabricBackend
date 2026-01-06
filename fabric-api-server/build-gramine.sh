#!/bin/bash
# Build script for fabric-api-server with Gramine SGX support
# This script builds the Go binary and prepares it for SGX execution

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
BINARY_NAME="fabric-api-server"
MANIFEST_TEMPLATE="fabric-api-server.manifest.toml"
MANIFEST_OUTPUT="fabric-api-server.manifest"
MANIFEST_SGX="fabric-api-server.manifest.sgx"
SIGNATURE_OUTPUT="fabric-api-server.sig"
SRC_DIR="./src"

echo -e "${GREEN}Building fabric-api-server for Gramine SGX...${NC}"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed or not in PATH${NC}"
    exit 1
fi

# Check if gramine-manifest is available
if ! command -v gramine-manifest &> /dev/null; then
    echo -e "${YELLOW}Warning: gramine-manifest not found. Make sure Gramine is installed.${NC}"
    echo -e "${YELLOW}You can still build the binary, but manifest generation will be skipped.${NC}"
    SKIP_MANIFEST=true
else
    SKIP_MANIFEST=false
fi

# Check if gramine-sgx-sign is available
if ! command -v gramine-sgx-sign &> /dev/null; then
    echo -e "${YELLOW}Warning: gramine-sgx-sign not found. Make sure Gramine SGX tools are installed.${NC}"
    echo -e "${YELLOW}Manifest signing will be skipped.${NC}"
    SKIP_SIGN=true
else
    SKIP_SIGN=false
fi

# Step 1: Build the Go binary
echo -e "${GREEN}Step 1: Building Go binary...${NC}"
echo "Building with: go build -o ${BINARY_NAME} ${SRC_DIR}/."

# Try static build first (recommended for SGX)
if CGO_ENABLED=0 go build -ldflags="-extldflags=-static" -o "${BINARY_NAME}" "${SRC_DIR}/." 2>/dev/null; then
    echo -e "${GREEN}✓ Built static binary: ${BINARY_NAME}${NC}"
    STATIC_BUILD=true
else
    echo -e "${YELLOW}Static build failed, trying regular build...${NC}"
    if go build -o "${BINARY_NAME}" "${SRC_DIR}/."; then
        echo -e "${GREEN}✓ Built binary: ${BINARY_NAME}${NC}"
        STATIC_BUILD=false
    else
        echo -e "${RED}Error: Failed to build Go binary${NC}"
        exit 1
    fi
fi

# Check if binary was created
if [ ! -f "${BINARY_NAME}" ]; then
    echo -e "${RED}Error: Binary ${BINARY_NAME} was not created${NC}"
    exit 1
fi

echo -e "${GREEN}Binary size: $(du -h ${BINARY_NAME} | cut -f1)${NC}"

# Step 2: Generate SGX manifest
if [ "$SKIP_MANIFEST" = false ]; then
    echo -e "${GREEN}Step 2: Generating SGX manifest...${NC}"
    
    if [ ! -f "${MANIFEST_TEMPLATE}" ]; then
        echo -e "${RED}Error: Manifest template ${MANIFEST_TEMPLATE} not found${NC}"
        exit 1
    fi
    
    # Generate manifest with gramine-manifest
    # The manifest template will be processed to create the final manifest
    gramine-manifest -Dlog_level=info "${MANIFEST_TEMPLATE}" "${MANIFEST_OUTPUT}"
    
    if [ -f "${MANIFEST_OUTPUT}" ]; then
        echo -e "${GREEN}✓ Generated manifest: ${MANIFEST_OUTPUT}${NC}"
    else
        echo -e "${RED}Error: Failed to generate manifest${NC}"
        exit 1
    fi
else
    echo -e "${YELLOW}Skipping manifest generation (gramine-manifest not available)${NC}"
fi

# Step 3: Sign the manifest
if [ "$SKIP_SIGN" = false ] && [ -f "${MANIFEST_OUTPUT}" ]; then
    echo -e "${GREEN}Step 3: Signing manifest for SGX...${NC}"
    
    # gramine-sgx-sign creates .manifest.sgx file
    gramine-sgx-sign --manifest "${MANIFEST_OUTPUT}" --output "${MANIFEST_SGX}"
    
    if [ -f "${MANIFEST_SGX}" ]; then
        echo -e "${GREEN}✓ Signed manifest: ${MANIFEST_SGX}${NC}"
    else
        echo -e "${RED}Error: Failed to sign manifest${NC}"
        exit 1
    fi
    
    if [ -f "${SIGNATURE_OUTPUT}" ]; then
        echo -e "${GREEN}✓ Signature file: ${SIGNATURE_OUTPUT}${NC}"
    fi
else
    if [ "$SKIP_SIGN" = true ]; then
        echo -e "${YELLOW}Skipping manifest signing (gramine-sgx-sign not available)${NC}"
    else
        echo -e "${YELLOW}Skipping manifest signing (manifest not generated)${NC}"
    fi
fi

# Summary
echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Build completed successfully!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "Files created:"
echo "  - ${BINARY_NAME} (Go binary)"
if [ -f "${MANIFEST_OUTPUT}" ]; then
    echo "  - ${MANIFEST_OUTPUT} (Gramine manifest)"
fi
if [ -f "${MANIFEST_SGX}" ]; then
    echo "  - ${MANIFEST_SGX} (SGX signed manifest)"
fi
if [ -f "${SIGNATURE_OUTPUT}" ]; then
    echo "  - ${SIGNATURE_OUTPUT} (SGX signature)"
fi
echo ""
echo "To run with Gramine (non-SGX mode for testing):"
echo "  gramine-direct ${BINARY_NAME} -p 8001"
echo ""
echo "To run with Gramine SGX:"
echo "  gramine-sgx ${BINARY_NAME} -p 8001"
echo ""
echo "Note: Make sure to set environment variables:"
echo "  export CRYPTO_PATH=/path/to/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com"
echo "  export PEER_ENDPOINT=localhost:7051"
echo "  export CHAINCODE_NAME=basic  # optional, default is 'basic'"
echo "  export CHANNEL_NAME=mychannel  # optional, default is 'mychannel'"
echo ""
