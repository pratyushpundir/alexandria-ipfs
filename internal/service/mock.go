package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"sync"
)

// base58 alphabet used by IPFS
const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// encodeBase58 encodes bytes to base58 string
func encodeBase58(input []byte) string {
	// Count leading zeros
	leadingZeros := 0
	for _, b := range input {
		if b != 0 {
			break
		}
		leadingZeros++
	}

	// Convert to big integer and encode
	// Simplified implementation for our mock purposes
	result := make([]byte, 0, len(input)*2)

	// Work with a copy to avoid modifying input
	data := make([]byte, len(input))
	copy(data, input)

	for len(data) > 0 {
		// Find first non-zero byte
		firstNonZero := 0
		for firstNonZero < len(data) && data[firstNonZero] == 0 {
			firstNonZero++
		}
		if firstNonZero == len(data) {
			break
		}
		data = data[firstNonZero:]

		// Divide by 58
		remainder := 0
		newData := make([]byte, 0, len(data))
		for _, b := range data {
			acc := remainder*256 + int(b)
			digit := acc / 58
			remainder = acc % 58
			if len(newData) > 0 || digit > 0 {
				newData = append(newData, byte(digit))
			}
		}
		data = newData
		result = append([]byte{base58Alphabet[remainder]}, result...)
	}

	// Add leading '1's for leading zeros in input
	for i := 0; i < leadingZeros; i++ {
		result = append([]byte{'1'}, result...)
	}

	return string(result)
}

// MockClient is a mock IPFS client for development without Blockfrost credentials
type MockClient struct {
	mu         sync.RWMutex
	storage    map[string][]byte
	pins       map[string]bool
	gatewayURL string
}

// NewMockClient creates a new mock IPFS client
func NewMockClient(gatewayURL string) *MockClient {
	if gatewayURL == "" {
		gatewayURL = "https://ipfs.io/ipfs"
	}
	log.Println("IPFS service running in MOCK MODE - data is stored in memory only")
	return &MockClient{
		storage:    make(map[string][]byte),
		pins:       make(map[string]bool),
		gatewayURL: gatewayURL,
	}
}

// Upload stores data in memory and returns a fake CID based on content hash
func (c *MockClient) Upload(ctx context.Context, data []byte, filename string) (*AddResponse, error) {
	// Generate a valid CIDv0 (base58-encoded multihash)
	// CIDv0 format: base58(multihash) where multihash = <hash-func-code><digest-length><digest>
	// For SHA2-256: hash-func-code = 0x12, digest-length = 0x20 (32 bytes)
	hash := sha256.Sum256(data)

	// Build multihash: 0x12 (sha2-256) + 0x20 (32 bytes) + hash
	multihash := make([]byte, 34)
	multihash[0] = 0x12 // sha2-256 function code
	multihash[1] = 0x20 // digest length (32 bytes)
	copy(multihash[2:], hash[:])

	// Encode as base58 to get valid CIDv0
	cid := encodeBase58(multihash)

	c.mu.Lock()
	c.storage[cid] = data
	c.mu.Unlock()

	log.Printf("[MOCK IPFS] Uploaded %d bytes as %s (filename: %s)", len(data), cid, filename)

	return &AddResponse{
		Name:     filename,
		IPFSHash: cid,
		Size:     fmt.Sprintf("%d", len(data)),
	}, nil
}

// Get retrieves content from memory
func (c *MockClient) Get(ctx context.Context, cid string) ([]byte, error) {
	c.mu.RLock()
	data, ok := c.storage[cid]
	c.mu.RUnlock()

	if !ok {
		// Content not found - this happens when mock storage was reset (container restart)
		// Return an error so the caller knows the content is unavailable
		log.Printf("[MOCK IPFS] CID %s not found in mock storage", cid)
		return nil, fmt.Errorf("content not found: %s (mock storage was likely reset)", cid)
	}

	log.Printf("[MOCK IPFS] Retrieved %d bytes for CID %s", len(data), cid)
	return data, nil
}

// Pin marks content as pinned in memory
func (c *MockClient) Pin(ctx context.Context, cid string) error {
	c.mu.Lock()
	c.pins[cid] = true
	c.mu.Unlock()

	log.Printf("[MOCK IPFS] Pinned CID %s", cid)
	return nil
}

// Unpin removes the pin from memory
func (c *MockClient) Unpin(ctx context.Context, cid string) error {
	c.mu.Lock()
	delete(c.pins, cid)
	c.mu.Unlock()

	log.Printf("[MOCK IPFS] Unpinned CID %s", cid)
	return nil
}

// GetGatewayURL returns a mock gateway URL
func (c *MockClient) GetGatewayURL(cid string) string {
	return fmt.Sprintf("%s/%s", c.gatewayURL, cid)
}

// IsPinned checks if content is pinned (for testing)
func (c *MockClient) IsPinned(cid string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.pins[cid]
}

// GetStoredCIDs returns all stored CIDs (for testing)
func (c *MockClient) GetStoredCIDs() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cids := make([]string, 0, len(c.storage))
	for cid := range c.storage {
		cids = append(cids, cid)
	}
	return cids
}

// Ensure MockClient implements IPFSClient
var _ IPFSClient = (*MockClient)(nil)
