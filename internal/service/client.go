package service

import "context"

// IPFSClient defines the interface for IPFS operations
type IPFSClient interface {
	Upload(ctx context.Context, data []byte, filename string) (*AddResponse, error)
	Get(ctx context.Context, cid string) ([]byte, error)
	Pin(ctx context.Context, cid string) error
	Unpin(ctx context.Context, cid string) error
	GetGatewayURL(cid string) string
}

// Ensure BlockfrostClient implements IPFSClient
var _ IPFSClient = (*BlockfrostClient)(nil)
