package aws

import (
	"context"
	"errors"
	"net/http"
	"testing"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
)

func TestNewLocalClientsConfiguresLocalStack(t *testing.T) {
	const endpointURL = "http://localstack:4566"

	clients, err := NewLocalClients(LocalClientOptions{
		Region:      "ap-southeast-1",
		EndpointURL: endpointURL,
	})
	if err != nil {
		t.Fatalf("NewLocalClients() unexpected error: %v", err)
	}
	if clients.EventBridge == nil || clients.DynamoDB == nil {
		t.Fatal("NewLocalClients() returned a nil service client")
	}

	eventBridgeOptions := clients.EventBridge.Options()
	if eventBridgeOptions.Region != "ap-southeast-1" {
		t.Errorf(
			"EventBridge region = %q, want %q",
			eventBridgeOptions.Region,
			"ap-southeast-1",
		)
	}
	if got := awssdk.ToString(eventBridgeOptions.BaseEndpoint); got != endpointURL {
		t.Errorf("EventBridge endpoint = %q, want %q", got, endpointURL)
	}
	if eventBridgeOptions.RetryMaxAttempts != localRetryAttempts {
		t.Errorf(
			"EventBridge retry attempts = %d, want %d",
			eventBridgeOptions.RetryMaxAttempts,
			localRetryAttempts,
		)
	}
	assertLocalHTTPClient(t, eventBridgeOptions.HTTPClient)
	if eventBridgeOptions.Credentials == nil {
		t.Error("EventBridge credentials provider is nil")
	} else {
		assertLocalCredentials(t, eventBridgeOptions.Credentials)
	}

	dynamoDBOptions := clients.DynamoDB.Options()
	if dynamoDBOptions.Region != "ap-southeast-1" {
		t.Errorf(
			"DynamoDB region = %q, want %q",
			dynamoDBOptions.Region,
			"ap-southeast-1",
		)
	}
	if got := awssdk.ToString(dynamoDBOptions.BaseEndpoint); got != endpointURL {
		t.Errorf("DynamoDB endpoint = %q, want %q", got, endpointURL)
	}
	if dynamoDBOptions.RetryMaxAttempts != localRetryAttempts {
		t.Errorf(
			"DynamoDB retry attempts = %d, want %d",
			dynamoDBOptions.RetryMaxAttempts,
			localRetryAttempts,
		)
	}
	assertLocalHTTPClient(t, dynamoDBOptions.HTTPClient)
	if dynamoDBOptions.Credentials == nil {
		t.Error("DynamoDB credentials provider is nil")
	}
}

func TestValidateLocalEndpointAllowsApprovedHosts(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		want     string
	}{
		{
			name:     "Docker service",
			endpoint: " http://localstack:4566/ ",
			want:     "http://localstack:4566",
		},
		{
			name:     "localhost",
			endpoint: "http://localhost:4566",
			want:     "http://localhost:4566",
		},
		{
			name:     "IPv4 loopback",
			endpoint: "http://127.0.0.2:4566",
			want:     "http://127.0.0.2:4566",
		},
		{
			name:     "IPv6 loopback",
			endpoint: "http://[::1]:4566",
			want:     "http://[::1]:4566",
		},
		{
			name:     "LocalStack localhost domain",
			endpoint: "https://edge.localhost.localstack.cloud:4566",
			want:     "https://edge.localhost.localstack.cloud:4566",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := validateLocalEndpoint(test.endpoint)
			if err != nil {
				t.Fatalf("validateLocalEndpoint() unexpected error: %v", err)
			}
			if got != test.want {
				t.Errorf("validateLocalEndpoint() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestNewLocalClientsRejectsUnsafeConfiguration(t *testing.T) {
	tests := []struct {
		name    string
		options LocalClientOptions
		wantErr error
	}{
		{
			name: "missing region",
			options: LocalClientOptions{
				EndpointURL: "http://localhost:4566",
			},
			wantErr: ErrRegionRequired,
		},
		{
			name: "missing endpoint",
			options: LocalClientOptions{
				Region: "ap-southeast-1",
			},
			wantErr: ErrLocalEndpointRequired,
		},
		{
			name: "AWS endpoint",
			options: LocalClientOptions{
				Region:      "ap-southeast-1",
				EndpointURL: "https://events.ap-southeast-1.amazonaws.com",
			},
			wantErr: ErrLocalEndpointRejected,
		},
		{
			name: "arbitrary remote host",
			options: LocalClientOptions{
				Region:      "ap-southeast-1",
				EndpointURL: "http://example.com:4566",
			},
			wantErr: ErrLocalEndpointRejected,
		},
		{
			name: "localhost lookalike",
			options: LocalClientOptions{
				Region:      "ap-southeast-1",
				EndpointURL: "http://localhost.evil:4566",
			},
			wantErr: ErrLocalEndpointRejected,
		},
		{
			name: "trailing-dot hostname",
			options: LocalClientOptions{
				Region:      "ap-southeast-1",
				EndpointURL: "http://localhost.:4566",
			},
			wantErr: ErrLocalEndpointRejected,
		},
		{
			name: "unsupported scheme",
			options: LocalClientOptions{
				Region:      "ap-southeast-1",
				EndpointURL: "ftp://localhost:4566",
			},
			wantErr: ErrLocalEndpointRejected,
		},
		{
			name: "credentials in URL",
			options: LocalClientOptions{
				Region:      "ap-southeast-1",
				EndpointURL: "http://user:secret@localhost:4566",
			},
			wantErr: ErrLocalEndpointRejected,
		},
		{
			name: "path in URL",
			options: LocalClientOptions{
				Region:      "ap-southeast-1",
				EndpointURL: "http://localhost:4566/events",
			},
			wantErr: ErrLocalEndpointRejected,
		},
		{
			name: "query in URL",
			options: LocalClientOptions{
				Region:      "ap-southeast-1",
				EndpointURL: "http://localhost:4566?target=aws",
			},
			wantErr: ErrLocalEndpointRejected,
		},
		{
			name: "fragment in URL",
			options: LocalClientOptions{
				Region:      "ap-southeast-1",
				EndpointURL: "http://localhost:4566#fragment",
			},
			wantErr: ErrLocalEndpointRejected,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewLocalClients(test.options)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("NewLocalClients() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func assertLocalHTTPClient(t *testing.T, client interface {
	Do(*http.Request) (*http.Response, error)
}) {
	t.Helper()

	httpClient, ok := client.(*http.Client)
	if !ok {
		t.Fatalf("HTTP client type = %T, want *http.Client", client)
	}
	if httpClient.Timeout != localHTTPTimeout {
		t.Errorf(
			"HTTP timeout = %v, want %v",
			httpClient.Timeout,
			localHTTPTimeout,
		)
	}
	if httpClient.CheckRedirect == nil {
		t.Error("HTTP redirect policy is nil")
	}

	transport, ok := httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("HTTP transport type = %T, want *http.Transport", httpClient.Transport)
	}
	if transport.Proxy != nil {
		t.Error("HTTP proxy configuration is enabled")
	}
}

func assertLocalCredentials(
	t *testing.T,
	provider awssdk.CredentialsProvider,
) {
	t.Helper()

	credentials, err := provider.Retrieve(context.Background())
	if err != nil {
		t.Fatalf("retrieve local credentials: %v", err)
	}
	if credentials.AccessKeyID != "localstack" {
		t.Errorf(
			"local access key ID = %q, want %q",
			credentials.AccessKeyID,
			"localstack",
		)
	}
	if credentials.SecretAccessKey != "localstack" {
		t.Errorf("local secret access key does not match the dummy value")
	}
	if credentials.SessionToken != "" {
		t.Errorf("local session token = %q, want empty", credentials.SessionToken)
	}
}
