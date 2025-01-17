// Package aws implements a koanf.Provider that takes a []byte
// and provides it to koanf to be parsed by a koanf.Parser.
package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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

	// SecretsManagerClient
	// If none is provided, the provider initialization will create one.
	SecretsManagerClient *secretsmanager.Client
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

	if cfg.SecretsManagerClient == nil {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
		defer cancel()

		awsCfg, err := config.LoadDefaultConfig(
			ctx,
			config.WithRegion(cfg.Region),
		)
		if err != nil {
			return nil, err
		}

		cfg.SecretsManagerClient = secretsmanager.NewFromConfig(awsCfg)
	}

	return &SecretsManager{
		cfg: cfg,
		svc: cfg.SecretsManagerClient,
	}, nil
}

// ReadBytes reads the contents of the secrt and returns the bytes.
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

	if resp.SecretString != nil {
		return []byte(*resp.SecretString), nil
	}

	return resp.SecretBinary, nil
}

// Read is not supported for the secrets manager provider.
func (p *SecretsManager) Read() (map[string]any, error) {
	return nil, fmt.Errorf("secretsmanager provider does not support this method")
}

type JSONParser struct {
	Prefix string
}

// Unmarshal parses the given JSON bytes.
func (p *JSONParser) Unmarshal(b []byte) (map[string]interface{}, error) {
	var out map[string]interface{}
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}

	for k, v := range out {
		if !strings.HasPrefix(k, p.Prefix) {
			m, ok := out[p.Prefix]
			if !ok {
				out[p.Prefix] = map[string]any{
					k: v,
				}
			}
			mm, ok := m.(map[string]any)
			if ok {
				mm[k] = v
			}

			delete(out, k)
		}
	}

	return out, nil
}

// Marshal marshals the given config map to JSON bytes.
func (p *JSONParser) Marshal(o map[string]interface{}) ([]byte, error) {
	for k, v := range o {
		if !strings.HasPrefix(k, p.Prefix) {
			m, ok := o[p.Prefix]
			if !ok {
				o[p.Prefix] = map[string]any{
					k: v,
				}
			}
			mm, ok := m.(map[string]any)
			if ok {
				mm[k] = v
			}

			delete(o, k)
		}
	}

	return json.Marshal(o)
}
