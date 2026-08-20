package aws

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
)

const (
	localHTTPTimeout   = 10 * time.Second
	localRetryAttempts = 3
)

var (
	ErrRegionRequired        = errors.New("AWS region is required")
	ErrLocalEndpointRequired = errors.New("LocalStack endpoint is required")
	ErrLocalEndpointRejected = errors.New("endpoint is not an approved local endpoint")
)

// LocalClientOptions configures AWS SDK clients that are restricted to
// LocalStack.
type LocalClientOptions struct {
	Region      string
	EndpointURL string
}

// LocalClients contains the AWS SDK clients needed by the current adapters.
type LocalClients struct {
	EventBridge *eventbridge.Client
	DynamoDB    *dynamodb.Client
}

// NewLocalClients constructs LocalStack-only service clients without making a
// network request. It deliberately has no production-AWS fallback.
func NewLocalClients(options LocalClientOptions) (LocalClients, error) {
	region := strings.TrimSpace(options.Region)
	if region == "" {
		return LocalClients{}, ErrRegionRequired
	}

	endpointURL, err := validateLocalEndpoint(options.EndpointURL)
	if err != nil {
		return LocalClients{}, err
	}

	credentialsProvider := awssdk.NewCredentialsCache(
		credentials.NewStaticCredentialsProvider("localstack", "localstack", ""),
	)
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   localHTTPTimeout,
		CheckRedirect: func(
			_ *http.Request,
			_ []*http.Request,
		) error {
			return http.ErrUseLastResponse
		},
	}

	eventBridgeClient := eventbridge.New(eventbridge.Options{
		Region:           region,
		BaseEndpoint:     awssdk.String(endpointURL),
		Credentials:      credentialsProvider,
		HTTPClient:       httpClient,
		RetryMaxAttempts: localRetryAttempts,
	})
	dynamoDBClient := dynamodb.New(dynamodb.Options{
		Region:           region,
		BaseEndpoint:     awssdk.String(endpointURL),
		Credentials:      credentialsProvider,
		HTTPClient:       httpClient,
		RetryMaxAttempts: localRetryAttempts,
	})

	return LocalClients{
		EventBridge: eventBridgeClient,
		DynamoDB:    dynamoDBClient,
	}, nil
}

func validateLocalEndpoint(rawEndpoint string) (string, error) {
	rawEndpoint = strings.TrimSpace(rawEndpoint)
	if rawEndpoint == "" {
		return "", ErrLocalEndpointRequired
	}

	parsed, err := url.Parse(rawEndpoint)
	if err != nil {
		return "", fmt.Errorf("%w: malformed URL", ErrLocalEndpointRejected)
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return "", fmt.Errorf("%w: HTTP or HTTPS is required", ErrLocalEndpointRejected)
	}
	if parsed.Opaque != "" || parsed.Host == "" || parsed.Hostname() == "" {
		return "", fmt.Errorf("%w: hostname is required", ErrLocalEndpointRejected)
	}
	if parsed.User != nil {
		return "", fmt.Errorf("%w: user information is not allowed", ErrLocalEndpointRejected)
	}
	if parsed.Path != "" && parsed.Path != "/" {
		return "", fmt.Errorf("%w: paths are not allowed", ErrLocalEndpointRejected)
	}
	if parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" {
		return "", fmt.Errorf("%w: query strings and fragments are not allowed", ErrLocalEndpointRejected)
	}
	if !isApprovedLocalHost(parsed.Hostname()) {
		return "", ErrLocalEndpointRejected
	}

	parsed.Path = ""
	return strings.TrimSuffix(parsed.String(), "/"), nil
}

func isApprovedLocalHost(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))

	switch host {
	case "localhost", "localstack", "host.docker.internal",
		"localhost.localstack.cloud":
		return true
	}

	if strings.HasSuffix(host, ".localhost.localstack.cloud") {
		return true
	}

	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
