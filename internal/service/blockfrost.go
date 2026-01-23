package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

// BlockfrostClient handles IPFS operations via Blockfrost API
type BlockfrostClient struct {
	projectID  string
	baseURL    string
	gatewayURL string
	httpClient *http.Client
}

// BlockfrostConfig holds Blockfrost client configuration
type BlockfrostConfig struct {
	ProjectID  string
	BaseURL    string
	GatewayURL string
	Timeout    time.Duration
}

// NewBlockfrostClient creates a new Blockfrost IPFS client
func NewBlockfrostClient(cfg *BlockfrostConfig) *BlockfrostClient {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 60 * time.Second
	}

	return &BlockfrostClient{
		projectID:  cfg.ProjectID,
		baseURL:    cfg.BaseURL,
		gatewayURL: cfg.GatewayURL,
		httpClient: &http.Client{Timeout: timeout},
	}
}

// AddResponse represents the response from Blockfrost IPFS add endpoint
type AddResponse struct {
	Name     string `json:"name"`
	IPFSHash string `json:"ipfs_hash"`
	Size     string `json:"size"`
}

// Upload uploads data to IPFS via Blockfrost
func (c *BlockfrostClient) Upload(ctx context.Context, data []byte, filename string) (*AddResponse, error) {
	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := part.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	// Create request
	url := fmt.Sprintf("%s/ipfs/add", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("project_id", c.projectID)

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to IPFS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("IPFS upload failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var addResp AddResponse
	if err := json.NewDecoder(resp.Body).Decode(&addResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &addResp, nil
}

// Get retrieves content from IPFS via Blockfrost gateway
func (c *BlockfrostClient) Get(ctx context.Context, cid string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", c.gatewayURL, cid)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get from IPFS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("IPFS get failed with status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// Pin pins content on Blockfrost IPFS
func (c *BlockfrostClient) Pin(ctx context.Context, cid string) error {
	url := fmt.Sprintf("%s/ipfs/pin/add/%s", c.baseURL, cid)

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("project_id", c.projectID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to pin: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("IPFS pin failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// Unpin removes a pin from Blockfrost IPFS
func (c *BlockfrostClient) Unpin(ctx context.Context, cid string) error {
	url := fmt.Sprintf("%s/ipfs/pin/remove/%s", c.baseURL, cid)

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("project_id", c.projectID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to unpin: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("IPFS unpin failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// GetGatewayURL returns the public gateway URL for a CID
func (c *BlockfrostClient) GetGatewayURL(cid string) string {
	return fmt.Sprintf("%s/%s", c.gatewayURL, cid)
}
