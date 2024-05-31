package aws

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/localstack"
)

func TestSecretsManager_ReadBytes(t *testing.T) {
	t.Run("container running and secret found", func(t *testing.T) {
		svc, cleanup := startContainer(t)
		t.Cleanup(cleanup)

		sm := &SecretsManager{
			cfg: Config{
				Secret:  "test",
				Timeout: _defaultTimeout,
			},
			svc: svc,
		}

		expected := []byte(`{"some":"json"}`)
		_, err := svc.CreateSecret(context.Background(), &secretsmanager.CreateSecretInput{
			Name:         aws.String(sm.cfg.Secret),
			SecretBinary: expected,
		})
		if err != nil {
			t.Fatal(err)
		}

		got, err := sm.ReadBytes()
		if err != nil {
			t.Error(err)
		}

		if !bytes.Equal(got, expected) {
			t.Errorf("bytes not equal, got %v, want %v", got, expected)
		}
	})

	t.Run("container not running", func(t *testing.T) {
		sm := &SecretsManager{
			cfg: Config{
				Secret: "test",
			},
			svc: secretsmanager.NewFromConfig(*aws.NewConfig()),
		}

		_, err := sm.ReadBytes()
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("secret missing", func(t *testing.T) {
		svc, cleanup := startContainer(t)
		t.Cleanup(cleanup)

		sm := &SecretsManager{
			cfg: Config{
				Secret: "test",
			},
			svc: svc,
		}

		_, err := sm.ReadBytes()
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func startContainer(t *testing.T) (*secretsmanager.Client, func()) {
	t.Helper()

	ctx := context.Background()
	container, err := localstack.RunContainer(
		ctx,
		testcontainers.WithImage("localstack/localstack:1.4.0"),
	)
	if err != nil {
		t.Fatal(err)
	}

	port, err := container.MappedPort(ctx, nat.Port("4566/tcp"))
	if err != nil {
		t.Fatal(err)
	}

	provider, err := testcontainers.NewDockerProvider()
	if err != nil {
		t.Fatal(err)
	}

	host, err := provider.DaemonHost(ctx)
	if err != nil {
		t.Fatal(err)
	}

	region := "us-east-1"
	resolver := aws.EndpointResolverWithOptionsFunc(
		func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           fmt.Sprintf("http://%s:%d", host, port.Int()),
				SigningRegion: region,
			}, nil
		},
	)

	awsCfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(region),
		config.WithEndpointResolverWithOptions(resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("accessKey", "secretKey", "token")),
	)
	if err != nil {
		t.Fatal(err)
	}

	svc := secretsmanager.NewFromConfig(awsCfg)

	return svc, func() {
		if err := provider.Close(); err != nil {
			t.Error(err)
		}
		if err := container.Terminate(context.Background()); err != nil {
			t.Error(err)
		}
	}
}
