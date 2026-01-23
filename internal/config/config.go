package config

import (
	"os"
	"strconv"
)

// Config holds the IPFS service configuration
type Config struct {
	GRPCPort              string
	BlockfrostProjectID   string
	BlockfrostIPFSBaseURL string
	IPFSGatewayURL        string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		GRPCPort:              getEnv("GRPC_PORT", "9093"),
		BlockfrostProjectID:   getEnv("BLOCKFROST_IPFS_PROJECT_ID", ""),
		BlockfrostIPFSBaseURL: getEnv("BLOCKFROST_IPFS_BASE_URL", "https://ipfs.blockfrost.io/api/v0"),
		IPFSGatewayURL:        getEnv("IPFS_GATEWAY_URL", "https://ipfs.blockfrost.dev/ipfs"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}
