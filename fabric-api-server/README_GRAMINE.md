# Gramine SGX Integration Guide

This guide explains how to build and run the `fabric-api-server` application within a Gramine SGX enclave for enhanced security.

## Overview

Gramine is a library OS that enables running unmodified applications in Intel SGX enclaves. This integration allows the fabric-api-server to run in a trusted execution environment, protecting sensitive operations like certificate handling and blockchain interactions.

## Prerequisites

1. **Gramine Installation**: Install Gramine and its dependencies
   ```bash
   # Follow the official Gramine installation guide:
   # https://gramine.readthedocs.io/en/latest/installation.html
   ```

2. **SGX Support**: 
   - For production: Intel SGX-capable hardware
   - For development/testing: SGX simulation mode or hardware

3. **Go Toolchain**: Go 1.24.3 or later

4. **Hyperledger Fabric Network**: A running Fabric network with the chaincode deployed

## Building for Gramine

### Step 1: Build the Application

Use the provided build script to compile the Go binary and generate the SGX manifest:

```bash
cd fabric-api-server
./build-gramine.sh
```

This script will:
1. Build the Go binary (`fabric-api-server`)
2. Generate the SGX manifest from the template
3. Sign the manifest for SGX execution

### Manual Build Process

If you prefer to build manually:

```bash
# Build the Go binary (static linking recommended for SGX)
CGO_ENABLED=0 go build -ldflags="-extldflags=-static" -o fabric-api-server ./src/.

# Generate the SGX manifest
gramine-manifest -Dlog_level=info fabric-api-server.manifest.toml fabric-api-server.manifest

# Sign the manifest
gramine-sgx-sign --manifest fabric-api-server.manifest --output fabric-api-server.sig
```

## Configuration

### Environment Variables

The application requires the following environment variables:

- **CRYPTO_PATH**: Path to Fabric crypto materials
  ```bash
  export CRYPTO_PATH=/path/to/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com
  ```

- **PEER_ENDPOINT**: Fabric peer gRPC endpoint
  ```bash
  export PEER_ENDPOINT=localhost:7051
  # Or use DNS format:
  export PEER_ENDPOINT=dns:///localhost:7051
  ```

- **CHAINCODE_NAME** (optional): Chaincode name (default: "basic")
  ```bash
  export CHAINCODE_NAME=basic
  ```

- **CHANNEL_NAME** (optional): Channel name (default: "mychannel")
  ```bash
  export CHANNEL_NAME=mychannel
  ```

- **TOKEN_FILE_PATH** (optional): Path to token storage file (default: "./tokens.json")
  ```bash
  export TOKEN_FILE_PATH=/path/to/tokens.json
  ```

- **--local-test or -l**: Run in local test mode without Fabric connection (for development)
  ```bash
  gramine-direct fabric-api-server -l -p 8001
  ```

### Certificate Path Configuration

The application reads certificates from the path specified in `CRYPTO_PATH`. The Gramine manifest mounts the root filesystem, so:

- **Absolute paths** work automatically (e.g., `/home/user/fabric-samples/...`)
- **Relative paths** depend on the working directory when running Gramine

Ensure that:
1. The `CRYPTO_PATH` directory contains:
   - `users/User1@org1.example.com/msp/signcerts/` - User certificate
   - `users/User1@org1.example.com/msp/keystore/` - Private key
   - `peers/peer0.org1.example.com/tls/ca.crt` - TLS certificate

2. The path is accessible from within the Gramine enclave (absolute paths are recommended)

## Running the Application

### Testing in Non-SGX Mode

Before running in SGX mode, test with Gramine in direct mode (non-SGX):

```bash
# Set environment variables
export CRYPTO_PATH=/path/to/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com
export PEER_ENDPOINT=localhost:7051

# Run in direct mode (no SGX)
gramine-direct fabric-api-server -p 8001
```

This allows you to verify that:
- Filesystem mounts work correctly
- Certificate files are accessible
- Network connections to Fabric peer succeed
- The application functions normally

### Running in SGX Mode

Once verified in direct mode, run with SGX protection:

```bash
# Set environment variables
export CRYPTO_PATH=/path/to/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com
export PEER_ENDPOINT=localhost:7051

# Run in SGX mode
gramine-sgx fabric-api-server -p 8001
```

### Running with Custom Port

The application accepts a port flag:

```bash
gramine-sgx fabric-api-server -p 8080
```

## Manifest Configuration

The manifest file (`fabric-api-server.manifest.toml`) contains SGX-specific settings:

### Key Settings

- **Enclave Size**: 512M (sufficient for Go runtime and application)
- **Thread Count**: 4 (adequate for REST API server)
- **Heap Size**: 256M
- **Stack Size**: 8M per thread

### Filesystem Mounts

The manifest mounts:
- Root filesystem (`/`) - for accessing certificate files
- System libraries (`/lib`, `/lib64`, `/usr`) - for Go runtime
- System configuration (`/etc`) - for TLS certificate validation
- Temporary filesystem (`/tmp`)
- System information (`/proc`, `/sys`)

### Network Configuration

The manifest enables:
- Outbound network connections (for gRPC to Fabric peer)
- Inbound network connections (for REST API server)

## Troubleshooting

### Certificate File Access Issues

**Problem**: Application cannot find certificate files

**Solutions**:
1. Use absolute paths in `CRYPTO_PATH`
2. Verify the path exists and is accessible
3. Check that the manifest has the root filesystem mounted
4. Ensure file permissions allow reading

### Network Connection Issues

**Problem**: Cannot connect to Fabric peer

**Solutions**:
1. Verify `PEER_ENDPOINT` is correct
2. Check that the Fabric network is running
3. Ensure network access is enabled in the manifest
4. For `localhost`, ensure the peer is accessible from the enclave

### Go Runtime Issues

**Problem**: Application fails to start or crashes

**Solutions**:
1. Use static linking: `CGO_ENABLED=0 go build ...`
2. Verify all required system libraries are mounted
3. Check Go version compatibility (requires Go 1.24.3+)
4. Review logs for specific error messages

### Manifest Generation Errors

**Problem**: `gramine-manifest` fails

**Solutions**:
1. Verify Gramine is properly installed
2. Check manifest syntax (TOML format)
3. Ensure all referenced files exist
4. Review Gramine documentation for syntax updates

## Security Considerations

1. **Remote Attestation**: For production, enable remote attestation in the manifest:
   ```toml
   sgx.remote_attestation = true
   ```

2. **Certificate Protection**: Certificates and private keys are protected within the SGX enclave

3. **Network Security**: Ensure TLS connections are properly configured

4. **Key Management**: Consider using secure key storage solutions for production deployments

## Performance Notes

- SGX enclaves have performance overhead compared to native execution
- Network I/O and file I/O may be slower due to enclave transitions
- Memory usage is constrained by the enclave size (512M in this configuration)

## Additional Resources

- [Gramine Documentation](https://gramine.readthedocs.io/)
- [Gramine Examples](https://github.com/gramineproject/examples)
- [Intel SGX Documentation](https://www.intel.com/content/www/us/en/developer/tools/software-guard-extensions/overview.html)
- [Hyperledger Fabric Documentation](https://hyperledger-fabric.readthedocs.io/)

## Support

For issues specific to this integration:
1. Check the troubleshooting section above
2. Review application logs
3. Test in non-SGX mode first to isolate issues
4. Consult Gramine and SGX documentation
