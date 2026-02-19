# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a REST API server for drone certificate and profile management on Hyperledger Fabric blockchain. It's a NASA drone certification system research project written in Go using the Huma v2 REST API framework.

## Common Commands

### Development

```bash
# Run in local test mode (no Fabric connection required)
go run ./src/. --local-test -p 8001
go run ./src/. -l -p 8001

# Run in production mode (requires running Fabric network)
go run ./src/. -p 8001

# Build binary
go build -o fabric-api-server ./src/.

# Static build (for SGX deployment)
CGO_ENABLED=0 go build -ldflags="-extldflags=-static" -o fabric-api-server ./src/.
```

### Docker

```bash
# Run with Docker Compose
docker compose up --build
```

### Gramine SGX

```bash
# Build for SGX (builds binary + generates manifest + signs)
./build-gramine.sh

# Run in non-SGX mode (testing)
gramine-direct fabric-api-server -p 8001

# Run in SGX mode
gramine-sgx fabric-api-server -p 8001
```

## Architecture

### Directory Structure

```
src/
├── main.go              # API endpoint definitions using Huma v2
└── utils/
    ├── fabric_apis.go   # Hyperledger Fabric blockchain interactions
    ├── record.go        # Record struct and utilities
    └── handlers.go      # Utility handlers (random string generation)
```

### Key Dependencies

- **Huma v2** (`github.com/danielgtaylor/huma/v2`) - REST API framework with OpenAPI docs
- **Chi Router** (`github.com/go-chi/chi/v5`) - HTTP router
- **Hyperledger Fabric Gateway** - Blockchain client SDK

### API Structure

The API uses Huma v2's declarative operation registration pattern:

```go
huma.Register(api, huma.Operation{
    OperationID:   "post-create-record",
    Method:        http.MethodPost,
    Path:          "/create-record",
    // ...
}, handlerFunc)
```

Interactive API documentation is auto-generated at `/docs` when the server is running.

### Data Model

The core `Record` struct (in `src/utils/record.go`) represents all blockchain records:

```go
type Record struct {
    RecordID   string `json:"recordID"`
    DroneID    string `json:"droneID"`
    PilotID    string `json:"pilotID"`
    ZoneID     string `json:"zoneID"`
    RecordType string `json:"recordType"`  // "certificate", "profile", etc.
    Reserved   string `json:"reserved"`     // Serialized JSON string for extensibility
}
```

### Blockchain Integration

The `fabric_apis.go` file handles all Fabric interactions:

- Uses the Fabric Gateway client SDK for gRPC communication
- Supports both `SubmitTransaction` (write) and `EvaluateTransaction` (query) operations
- All functions accept a `localTest` boolean to bypass blockchain calls during development

Connection is initialized via `utils.InitGateway()` which reads:
- **CRYPTO_PATH**: Path to Fabric crypto materials
- **PEER_ENDPOINT**: Fabric peer gRPC endpoint
- **CHAINCODE_NAME**: Chaincode name (default: "basic")
- **CHANNEL_NAME**: Channel name (default: "mychannel")

### Local Test Mode

When `--local-test` flag is set:
- No Fabric connection is established
- Blockchain operations are logged to console but not executed
- Useful for API development without a running Fabric network

### SGX Support

The application supports running in Intel SGX enclaves via Gramine:
- Static linking recommended (`CGO_ENABLED=0`)
- Manifest template at `fabric-api-server.manifest.toml`
- Build script handles binary compilation, manifest generation, and signing
