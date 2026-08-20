// Package config reads optional runtime settings for GameOps Lab.
package config

import "os"

// Config contains the values used by future local AWS-compatible adapters.
// All values may be empty during the repository-foundation phase.
type Config struct {
	AWSRegion         string
	AWSEndpointURL    string
	EventBusName      string
	DynamoDBTableName string
}

// Load reads configuration from environment variables without requiring them.
func Load() Config {
	return Config{
		AWSRegion:         os.Getenv("AWS_REGION"),
		AWSEndpointURL:    os.Getenv("AWS_ENDPOINT_URL"),
		EventBusName:      os.Getenv("EVENT_BUS_NAME"),
		DynamoDBTableName: os.Getenv("DYNAMODB_TABLE_NAME"),
	}
}
