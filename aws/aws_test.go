package aws

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/knadh/koanf/v2"
)

func TestLoadWithJSONParser(t *testing.T) {
	t.Setenv("LOCALSTACK_ENDPOINT", "http://localhost:4566")
	svc := newTestClient(t)

	t.Run("success", func(t *testing.T) {
		secretValue := []byte(`{
			"api_key": "testapikey"
		}`)
		secretName := createTestSecret(t, svc, secretValue)

		provider, err := Provider(Config{Secret: secretName, SecretsManagerClient: svc})
		if err != nil {
			t.Fatal(err)
		}

		k := koanf.New(".")

		if err := k.Load(provider, &JSONParser{Prefix: "stripe"}); err != nil {
			t.Fatal(err)
		}

		if k.Get("stripe.api_key") != "testapikey" {
			t.Errorf("LoadWithJSONParser() got = %v, want %v", k.Get("stripe.api_key"), "testapikey")
		}
	})

	t.Run("unmarshal", func(t *testing.T) {
		secretValue := []byte(`{
			"api_key": "testapikey"
		}`)
		secretName := createTestSecret(t, svc, secretValue)

		provider, err := Provider(Config{Secret: secretName, SecretsManagerClient: svc})
		if err != nil {
			t.Fatal(err)
		}

		k := koanf.New(".")

		if err := k.Load(provider, &JSONParser{Prefix: "stripe"}); err != nil {
			t.Fatal(err)
		}

		if k.Get("stripe.api_key") != "testapikey" {
			t.Errorf("LoadWithJSONParser() got = %v, want %v", k.Get("stripe.api_key"), "testapikey")
		}

		type Conf struct {
			Stripe struct {
				APIKey string `koanf:"api_key"`
			} `koanf:"stripe"`
		}

		var c Conf
		if err := k.UnmarshalWithConf("", &c, koanf.UnmarshalConf{Tag: "koanf"}); err != nil {
			t.Fatal(err)
		}

		if c.Stripe.APIKey != "testapikey" {
			t.Errorf("Unmarshaled conf got = %v, want %v", c.Stripe.APIKey, "testapikey")
		}
	})
}

func TestSecretsManager_ReadBytes(t *testing.T) {
	t.Setenv("LOCALSTACK_ENDPOINT", "http://localhost:4566")
	svc := newTestClient(t)

	t.Run("container running and secret found", func(t *testing.T) {
		expectedValue := []byte(`{"some":"json"}`)
		secretName := createTestSecret(t, svc, expectedValue)

		sm := &SecretsManager{
			cfg: Config{
				Secret:  secretName,
				Timeout: _defaultTimeout,
			},
			svc: svc,
		}

		got, err := sm.ReadBytes()
		if err != nil {
			t.Error(err)
		}

		if !bytes.Equal(got, expectedValue) {
			t.Errorf("bytes not equal, got %v, want %v", got, expectedValue)
		}
	})

	t.Run("secret missing", func(t *testing.T) {
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

func newTestClient(t *testing.T) *secretsmanager.Client {
	t.Helper()

	endpoint := os.Getenv("LOCALSTACK_ENDPOINT")
	if endpoint == "" {
		t.Skip("LOCALSTACK_ENDPOINT not set; run `docker compose up -d --wait` first")
	}

	awsCfg, err := config.LoadDefaultConfig(
		t.Context(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("accessKey", "secretKey", "token")),
	)
	if err != nil {
		t.Fatal(err)
	}

	return secretsmanager.NewFromConfig(awsCfg, func(o *secretsmanager.Options) {
		o.BaseEndpoint = &endpoint
	})
}

func createTestSecret(t *testing.T, svc *secretsmanager.Client, value []byte) string {
	t.Helper()

	name := strings.NewReplacer("/", "-", " ", "_").Replace(t.Name())
	_, err := svc.CreateSecret(t.Context(), &secretsmanager.CreateSecretInput{
		Name:         &name,
		SecretBinary: value,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_, err = svc.DeleteSecret(context.Background(), &secretsmanager.DeleteSecretInput{
			SecretId:                   &name,
			ForceDeleteWithoutRecovery: new(true),
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	return name
}
