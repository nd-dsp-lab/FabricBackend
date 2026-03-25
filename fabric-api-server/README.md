# fabric-api-server

A REST API server for drone certificate and profile management on Hyperledger Fabric blockchain. This Go-based server provides a RESTful interface for interacting with a Hyperledger Fabric network to create, query, and manage records stored on the blockchain.

## Features

- **RESTful API** for blockchain operations using [Huma](https://huma.rocks/) framework
- **Certificate Management** - Create and query drone pilot certificates
- **Profile Management** - Create, query, and update pilot/drone profiles
- **Record History Tracking** - Query the complete history of any record
- **Local Test Mode** - Run without a Fabric connection for development and testing
- **Docker Support** - Containerized deployment with Docker Compose

## Prerequisites

- **Go 1.24+** (for local development)
- **Hyperledger Fabric test network** (for production mode)
  - Requires a running Fabric network with chaincode deployed
  - See [Hyperledger Fabric Samples](https://github.com/hyperledger/fabric-samples) for setup instructions
- **Docker and Docker Compose** (for containerized deployment)

## Configuration

The server can be configured using the following environment variables:

| Variable | Description | Default Value |
|----------|-------------|---------------|
| `CRYPTO_PATH` | Path to Fabric crypto materials | `../fabric-samples/test-network/organizations/peerOrganizations/org1.example.com` |
| `PEER_ENDPOINT` | Fabric peer endpoint | `localhost:7051` |
| `CHAINCODE_NAME` | Chaincode name | `basic` |
| `CHANNEL_NAME` | Channel name | `mychannel` |

## Usage

### Local Development Mode (No Fabric Required)

The local test mode allows you to run the server without a Hyperledger Fabric connection. In this mode, all blockchain operations are logged to the console but not actually executed.

```bash
go run ./src/. --local-test -p 8001
# or
go run ./src/. -l -p 8001
```

### Production Mode (Requires Fabric Network)

```bash
# Ensure your Fabric network is running first
go run ./src/. -p 8001
```

### Docker Deployment

```bash
docker compose up --build
```

The API will be available at `http://localhost:8001`. Interactive API documentation is available at `http://localhost:8001/docs`.

## CLI Options

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--port` | `-p` | Port to listen on | `8001` |
| `--local-test` | `-l` | Run in local test mode without Fabric connection | `false` |

## Authentication

Some endpoints require Bearer token authentication. The default token is seeded for local testing:

| Token | Secret Key |
|-------|------------|
| `dev-test-token` | `dev-secret-key` |

### Token Management

Tokens are stored in a JSON file. The default location is `./tokens.json` in the working directory.

**Environment Variables:**

| Variable | Description | Default |
|----------|-------------|---------|
| `TOKEN_FILE_PATH` | Path to token storage file | `./tokens.json` |

**Token File Format:**
```json
{
  "tokens": {
    "your-token": "your-secret-key"
  }
}
```

### Authenticated Endpoints

The following endpoints require Bearer token authentication:

- `POST /create-record`
- `POST /certificates/create`
- `POST /profiles/create`

**Example:**
```bash
curl -X POST http://localhost:8001/certificates/create \
  -H "Authorization: Bearer dev-test-token" \
  -H "X-Secret-Key: dev-secret-key" \
  -H "Content-Type: application/json" \
  -d '{
    "droneID": "drone1",
    "pilotID": "Pilot00",
    "zoneID": "Zone00",
    "reserved": "{\"expiry\": \"2025-12-31\"}"
  }'
```

## Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `CRYPTO_PATH` | Path to Fabric crypto materials | `../fabric-samples/test-network/organizations/peerOrganizations/org1.example.com` |
| `PEER_ENDPOINT` | Fabric peer endpoint | `localhost:7051` |
| `CHAINCODE_NAME` | Chaincode name | `basic` |
| `CHANNEL_NAME` | Channel name | `mychannel` |
| `TOKEN_FILE_PATH` | Path to token storage file | `./tokens.json` |

## API Endpoints

All endpoints are relative to the base URL `http://localhost:8001`.

### Records

#### Create a Record (Authenticated)
```http
POST /create-record
```

Create a generic record on the blockchain.

**Request Body:**
```json
{
  "recordID": "record-001",
  "droneID": "drone1",
  "pilotID": "Pilot00",
  "zoneID": "Zone00",
  "recordType": "flight-log",
  "reserved": "{\"key\": \"value\"}"
}
```

**Response:**
```json
{
  "message": "Record created successfully.",
  "record": {
    "recordID": "record-001",
    "droneID": "drone1",
    "pilotID": "Pilot00",
    "zoneID": "Zone00",
    "recordType": "flight-log",
    "reserved": "{\"key\": \"value\"}"
  }
}
```

#### Query Records with Selector
```http
POST /records/query
```

Query records using a CouchDB Mango selector.

**Request Body:**
```json
{
  "selector": {
    "droneID": "drone1",
    "recordType": "flight-log"
  }
}
```

**Response:**
```json
{
  "body": {
    "message": "Found 2 records",
    "records": [
      {
        "recordID": "record-001",
        "droneID": "drone1",
        "pilotID": "Pilot00",
        "zoneID": "Zone00",
        "recordType": "flight-log",
        "reserved": "{}"
      }
    ]
  }
}
```

#### Get All Records
```http
GET /records/all
```

Retrieve all records from the blockchain.

**Response:**
```json
{
  "message": "Found 10 records",
  "records": [...]
}
```

#### Update a Record
```http
POST /records/update
```

Update an existing record.

**Request Body:**
```json
{
  "recordID": "record-001",
  "droneID": "drone1",
  "pilotID": "Pilot00",
  "zoneID": "Zone01",
  "recordType": "flight-log",
  "reserved": "{\"status\": \"completed\"}"
}
```

#### Get Record History
```http
POST /records/history
```

Get the complete history of a record including all previous versions.

**Request Body:**
```json
{
  "recordID": "record-001"
}
```

**Response:**
```json
{
  "message": "Found 3 history records",
  "history": [
    {
      "txId": "abc123...",
      "timestamp": "2024-01-15T10:30:00Z",
      "isDelete": false,
      "record": {...}
    }
  ]
}
```

### Certificates

#### Create a Certificate (Authenticated)
```http
POST /certificates/create
```

Create a certificate record. If `recordID` is not provided, a random one will be generated.

**Request Body:**
```json
{
  "recordID": "cert-001",
  "droneID": "drone1",
  "pilotID": "Pilot00",
  "zoneID": "Zone00",
  "reserved": "{\"expiry\": \"2025-12-31\"}"
}
```

**Response:**
```json
{
  "message": "Certificate received and stored successfully.",
  "record": {
    "recordID": "cert-001",
    "droneID": "drone1",
    "pilotID": "Pilot00",
    "zoneID": "Zone00",
    "recordType": "certificate",
    "reserved": "{\"expiry\": \"2025-12-31\"}"
  }
}
```

#### Query Certificates
```http
POST /certificates/query
```

Query certificates using a Mango selector. The selector automatically filters for `recordType: "certificate"`.

**Request Body:**
```json
{
  "selector": {
    "recordType": "certificate",
    "pilotID": "Pilot00"
  }
}
```

### Profiles

#### Create a Profile (Authenticated)
```http
POST /profiles/create
```

Create a profile record. If `recordID` is not provided, a random one will be generated.

**Request Body:**
```json
{
  "recordID": "profile-001",
  "droneID": "drone1",
  "pilotID": "Pilot00",
  "zoneID": "Zone00",
  "reserved": "{\"name\": \"John Doe\"}"
}
```

#### Query Profiles
```http
POST /profiles/query
```

Query profiles using a Mango selector. The selector automatically filters for `recordType: "profile"`.

**Request Body:**
```json
{
  "selector": {
    "recordType": "profile",
    "pilotID": "Pilot00"
  }
}
```

#### Update a Profile
```http
POST /profiles/update
```

Update an existing profile. Only fields provided in `updateInfo` will be updated.

**Request Body:**
```json
{
  "recordID": "profile-001",
  "updateInfo": {
    "droneID": "drone2",
    "pilotID": "Pilot01",
    "zoneID": "Zone01",
    "reserved": "{\"name\": \"Jane Doe\"}"
  }
}
```

## Examples

### Create a Certificate

```bash
curl -X POST http://localhost:8001/certificates/create \
  -H "Content-Type: application/json" \
  -d '{
    "droneID": "drone1",
    "pilotID": "Pilot00",
    "zoneID": "Zone00",
    "reserved": "{\"expiry\": \"2025-12-31\"}"
  }'
```

### Query Certificates by Pilot

```bash
curl -X POST http://localhost:8001/certificates/query \
  -H "Content-Type: application/json" \
  -d '{
    "selector": {
      "pilotID": "Pilot00"
    }
  }'
```

### Create a Profile

```bash
curl -X POST http://localhost:8001/profiles/create \
  -H "Content-Type: application/json" \
  -d '{
    "droneID": "drone1",
    "pilotID": "Pilot00",
    "zoneID": "Zone00",
    "reserved": "{\"name\": \"John Doe\", \"license\": \"UAV-12345\"}"
  }'
```

### Update a Profile

```bash
curl -X POST http://localhost:8001/profiles/update \
  -H "Content-Type: application/json" \
  -d '{
    "recordID": "profile-001",
    "updateInfo": {
      "zoneID": "Zone01",
      "reserved": "{\"name\": \"John Doe\", \"license\": \"UAV-12345\", \"status\": \"active\"}"
    }
  }'
```

### Get Record History

```bash
curl -X POST http://localhost:8001/records/history \
  -H "Content-Type: application/json" \
  -d '{
    "recordID": "record-001"
  }'
```

## Project Structure

```
fabric-api-server/
├── src/
│   ├── main.go              # Main application with API endpoints
│   └── utils/
│       └── fabric_apis.go   # Fabric blockchain interaction functions
├── go.mod                   # Go module definition
├── go.sum                   # Go dependencies checksum
├── Dockerfile               # Docker image definition
├── docker-compose.yml       # Docker Compose configuration
└── README.md                # This file
```

## Dependencies

- [Huma v2](https://huma.rocks/) - Modern REST API framework for Go
- [Chi Router](https://github.com/go-chi/chi) - Lightweight HTTP router
- [Hyperledger Fabric Gateway](https://github.com/hyperledger/fabric-gateway) - Fabric client SDK
- [gRPC](https://google.golang.org/grpc) - RPC framework

## License

This project is part of the NASA drone certification system research project.
