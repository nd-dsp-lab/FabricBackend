# AGENTS.md - FabricBackend Development Guide

This file provides guidance for AI agents working in this repository.

## Project Overview

FabricBackend is a Hyperledger Fabric blockchain backend for drone certification. It consists of:
- **fabric-api-server**: REST API server (Go, Huma v2 framework)
- **drp-storage**: Hyperledger Fabric chaincode (Go)

## Build Commands

### fabric-api-server
```bash
# Run in local test mode (no Fabric required)
go run ./fabric-api-server/src/. --local-test -p 8001
go run ./fabric-api-server/src/. -l -p 8001

# Run in production mode (requires Fabric network)
go run ./fabric-api-server/src/. -p 8001

# Build binary
go build -o fabric-api-server ./fabric-api-server/src/.

# Static build (for SGX/Gramine deployment)
CGO_ENABLED=0 go build -ldflags="-extldflags=-static" -o fabric-api-server ./fabric-api-server/src/.

# Docker
docker compose -f fabric-api-server/docker-compose.yml up --build
```

### drp-storage (Chaincode)
```bash
cd drp-storage/chaincode-go

# Build
go build

# Run tests
go test -v ./...

# Run single test
go test -v -run TestName ./...

# Lint
go vet ./...
gofmt -d .
```

### Makefile (Root level)
```bash
make check-prerequisite   # Check prerequisites
make install             # Install Fabric binaries
make drp_couchdb_deploy  # Bring up network and deploy chaincode
make api_server          # Start API server
make down                # Bring network down
make clean               # Clean up downloaded repos
```

## Code Style Guidelines

### Go Conventions

1. **Imports**: Use grouped imports (standard library first, then third-party)
   ```go
   import (
       "context"
       "fmt"
       "time"

       "github.com/danielgtaylor/huma/v2"
       "go-huma-api-server/src/utils"
   )
   ```

2. **Formatting**: Run `gofmt` before committing. Use 4-space indentation (Go standard).

3. **Naming**:
   - Use PascalCase for exported functions/types/structs
   - Use camelCase for unexported names
   - Use descriptive names: `RecordID`, `DroneID`, not `rid`, `did`
   - Acronyms: `URL` (all caps), but `url` for unexported

4. **Types**:
   - Use structs with JSON tags for API request/response bodies
   - Use interfaces for dependency injection (e.g., Fabric API abstraction)
   - Prefer concrete types unless polymorphism needed

5. **Error Handling**:
   - Return errors explicitly, don't ignore with `_`
   - Wrap errors with context: `fmt.Errorf("failed to X: %w", err)`
   - Use sentinel errors for known conditions
   - Log errors appropriately: `log.Printf("Warning: %v", err)`

### Project-Specific Patterns

6. **Huma v2 API Registration**:
   ```go
   huma.Register(api, huma.Operation{
       OperationID: "post-create-record",
       Method:       http.MethodPost,
       Path:         "/create-record",
       Description:  "Create a new record",
   }, handlerFunc)
   ```

7. **Struct Tags**: Use `json:"fieldName"` and `doc:"Description"` tags for OpenAPI docs
   ```go
   type RecordInput struct {
       Body struct {
           RecordID string `json:"recordID" doc:"Unique record ID"`
       }
   }
   ```

8. **Local Test Mode**: Functions that interact with Fabric should accept `localTest bool` parameter to bypass blockchain calls during development.

9. **Blockchain Interaction**:
   - Use `SubmitTransaction` for write operations
   - Use `EvaluateTransaction` for read operations
   - Handle connection lifecycle: `InitGateway()`, `defer Close()`

10. **Configuration**: Use environment variables for deployment-specific settings:
    - `CRYPTO_PATH`: Path to Fabric crypto materials
    - `PEER_ENDPOINT`: Fabric peer endpoint
    - `CHAINCODE_NAME`: Chaincode name
    - `CHANNEL_NAME`: Channel name

### Chaincode (drp-storage)

11. Chaincode follows Hyperledger Fabric conventions
12. Use `shim.ChaincodeStubInterface` for state operations
13. Implement `Init` and `Invoke` functions
14. Tests use testify for assertions

## Testing Guidelines

- Tests go in `*_test.go` files in the same package
- Use table-driven tests for multiple test cases
- Mock external dependencies (Fabric stub, etc.)
- Run tests before committing: `go test ./...`

## File Organization

```
fabric-api-server/
├── src/
│   ├── main.go          # API endpoints
│   ├── auth/            # Authentication (middleware, tokens)
│   └── utils/            # Fabric APIs, record types, handlers
├── go.mod               # Go 1.24+
└── Dockerfile

drp-storage/
├── chaincode-go/
│   ├── drp_storage.go   # Chaincode implementation
│   └── drp_storage_test.go
└── go.mod               # Go 1.22.5+
```

## Dependencies

- **fabric-api-server**: Huma v2, Chi router, Fabric Gateway SDK
- **drp-storage**: Fabric chaincode-go, Fabric contract-api-go, testify

## Notes

- Interactive API docs available at `/docs` when server running
- Default API port: 8001 (fabric-api-server)
- Default port: 6999 (old drp-client, deprecated)