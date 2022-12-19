package main

import (
	"context"
	"fmt"
	"testing"

	. "authentication/proto"

	. "github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

var (
	testAddr string          = "127.0.0.1:20001"
	testCtx  context.Context = context.TODO()
)

func TestClient(t *testing.T) {
	var (
		err    error
		conn   *grpc.ClientConn
		client AuthServiceClient

		in  *CreateQ
		ans *CreateA
	)

	conn, err = grpc.Dial(testAddr,
		grpc.WithInsecure(),
		// grpc.WithUnaryInterceptor(?grpc.UnaryClientInterceptor),
		// grpc.WithStreamInterceptor(?grpc.StreamClientInterceptor),
	)
	NoError(t, err)

	client = NewAuthServiceClient(conn)
	in = &CreateQ{Password: "123456"}

	ans, err = client.Create(testCtx, in)
	NoError(t, err)
	fmt.Printf("~~~ %+v\n", ans)
}
