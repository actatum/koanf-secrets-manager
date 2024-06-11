package gcp

import (
	"context"
	"net"
	"testing"

	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestSecretsManager_ReadBytes(t *testing.T) {
	t.Run("", func(t *testing.T) {
		l, err := net.Listen("tcp", "localhost:0")
		if err != nil {
			t.Fatal(err)
		}
		grpcServer := grpc.NewServer()
		secretmanagerpb.RegisterSecretManagerServiceServer(grpcServer, &fakeServer{})
		go func() {
			if err := grpcServer.Serve(l); err != nil {
				panic(err)
			}
		}()

		conn, err := grpc.Dial(l.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			t.Fatal(err)
		}

		client := secretmanagerpb.NewSecretManagerServiceClient(conn)
		sm := &SecretsManager{
			client: client,
		}
	})
}

type fakeServer struct {
	secretmanagerpb.UnimplementedSecretManagerServiceServer
}

func (f *fakeServer) AccessSecretVersion(ctx context.Context, in *secretmanagerpb.AccessSecretVersionRequest) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	return nil, nil
}
