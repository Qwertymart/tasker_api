package transport

import (
	"context"
	"google.golang.org/grpc/credentials/insecure"
	"task/pkg/userpb"
	"time"

	"google.golang.org/grpc"
)

type UserClient struct {
	client userpb.UserServiceClient
	conn   *grpc.ClientConn
}

func NewUserClient(address string) (*UserClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // Without TLS
	)
	if err != nil {
		return nil, err
	}
	client := userpb.NewUserServiceClient(conn)
	return &UserClient{client: client, conn: conn}, nil
}

// Closing the connection
func (uc *UserClient) Close() {
	err := uc.conn.Close()
	if err != nil {
		return
	}
}

func (uc *UserClient) CheckUser(ctx context.Context, id string) (bool, error) {
	// We are waiting for an answer 3 seconds
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	resp, err := uc.client.GetUser(ctx, &userpb.GetUserRequest{Id: id})
	if err != nil {
		return false, err
	}
	return resp.Exists, nil
}
