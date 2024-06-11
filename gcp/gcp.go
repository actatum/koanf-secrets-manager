package gcp

import (
	"context"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

const (
	_defaultTimeout = 5 * time.Second
)

// Config for the provider.
type Config struct {
	// GCP Project ID
	ProjectID string

	// Secret Name
	Secret string

	// Secret Version
	Version string

	// Timeout determines the context timeout for gcp requests.
	Timeout time.Duration
}

// SecretsManager implements a secrets manager provider.
type SecretsManager struct {
	cfg    Config
	client *secretmanager.Client
}

// Provider returns a provider from the given config.
func Provider(cfg Config) (*SecretsManager, error) {
	if cfg.Timeout == 0 {
		cfg.Timeout = _defaultTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	return &SecretsManager{
		cfg:    cfg,
		client: client,
	}, nil
}

// ReadBytes reads the contents of the secret and returns the bytes.
func (p *SecretsManager) ReadBytes() ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.cfg.Timeout)
	defer cancel()

	resp, err := p.client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: "",
	})
	if err != nil {
		return nil, err
	}

	return resp.Payload.Data, nil
}
