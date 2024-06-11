// Package aws implements a koanf.Provider that takes a []byte
// and provides it to koanf to be parsed by a koanf.Parser.
package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

const (
	_defaultTimeout = 5 * time.Second
)

// Config for the provider.
type Config struct {
	// AWS Region
	Region string

	// Secret Name
	Secret string

	// Secret Version
	Version string

	// Timeout determines the context timeout for aws requests.
	Timeout time.Duration
}

// SecretsManager implements a secrets manager provider.
type SecretsManager struct {
	cfg Config
	svc *secretsmanager.Client
}

// Provider returns a provider from the given config.
func Provider(cfg Config) (*SecretsManager, error) {
	if cfg.Timeout == 0 {
		cfg.Timeout = _defaultTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	awsCfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(cfg.Region),
	)
	if err != nil {
		return nil, err
	}

	svc := secretsmanager.NewFromConfig(awsCfg)

	return &SecretsManager{
		cfg: cfg,
		svc: svc,
	}, nil
}

// ReadBytes reads the contents of the secret and returns the bytes.
func (p *SecretsManager) ReadBytes() ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.cfg.Timeout)
	defer cancel()

	version := p.cfg.Version
	if version == "" {
		version = "AWSCURRENT"
	}

	resp, err := p.svc.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(p.cfg.Secret),
		VersionStage: aws.String(version),
	})
	if err != nil {
		return nil, err
	}

	return resp.SecretBinary, nil
}

// Read is not supported for the secrets manager provider.
func (p *SecretsManager) Read() (map[string]any, error) {
	return nil, fmt.Errorf("secretsmanager provider does not support this method")
}
