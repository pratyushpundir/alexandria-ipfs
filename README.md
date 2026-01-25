# Alexandria IPFS Service

Go gRPC microservice for IPFS content storage via Blockfrost.

## Features

- IPFS content pinning via Blockfrost API
- JSON metadata upload for NFTs
- Content retrieval by CID
- Health checks with gRPC health probe

## Tech Stack

- Go 1.25
- gRPC
- Blockfrost IPFS API

## Getting Started

### Prerequisites

- Go 1.25+
- Blockfrost API key with IPFS access

### Environment Variables

```bash
BLOCKFROST_API_KEY=your-blockfrost-ipfs-key
BLOCKFROST_IPFS_URL=https://ipfs.blockfrost.io/api/v0
GRPC_PORT=9092
```

### Running Locally

```bash
# Install dependencies
go mod download

# Start the server
go run ./cmd/server

# Or with Docker
docker compose up ipfs
```

## Project Structure

```
ipfs/
├── cmd/server/           # Entry point
├── internal/
│   ├── server/           # gRPC server implementation
│   └── blockfrost/       # Blockfrost IPFS client
├── proto/ipfs/v1/        # Protobuf definitions
├── gen/proto/            # Generated code
└── Dockerfile
```

## gRPC API

| Method | Description |
|--------|-------------|
| `PinContent` | Pin raw content to IPFS |
| `PinJSON` | Pin JSON metadata to IPFS |
| `GetContent` | Retrieve content by CID |

### Example: Pin JSON Metadata

```go
client := ipfspb.NewIPFSServiceClient(conn)

resp, err := client.PinJSON(ctx, &ipfspb.PinJSONRequest{
    Name: "credential-metadata",
    Content: `{
        "name": "Course Completion",
        "description": "Completed Introduction to Blockchain",
        "image": "ipfs://Qm..."
    }`,
})
// resp.IpfsHash: "Qm..."
```

## Blockfrost Integration

Uses Blockfrost's IPFS API for pinning:
- Automatic pinning on upload
- Content addressed storage
- Gateway access via `ipfs.blockfrost.io`

### IPFS Gateway URLs

```
https://ipfs.blockfrost.io/ipfs/{cid}
```

## Docker

```bash
# Build image
docker build -t alexandria-ipfs .

# Run container
docker run -p 9092:9092 \
  -e BLOCKFROST_API_KEY=key \
  alexandria-ipfs
```

## Ports

| Port | Protocol | Description |
|------|----------|-------------|
| 9092 | gRPC | IPFS service |

## Health Checks

Implements gRPC health checking protocol:

```bash
grpc_health_probe -addr=localhost:9092
```
