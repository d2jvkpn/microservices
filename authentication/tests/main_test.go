package tests

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

		cIn  *CreateQ
		cAns *CreateA
		vIn  *VerifyQ
		vAns *VerifyA
	)

	conn, err = grpc.Dial(testAddr,
		grpc.WithInsecure(),
		// grpc.WithUnaryInterceptor(?grpc.UnaryClientInterceptor),
		// grpc.WithStreamInterceptor(?grpc.StreamClientInterceptor),
	)
	NoError(t, err)

	client = NewAuthServiceClient(conn)
	cIn = &CreateQ{Password: "123456"}

	cAns, err = client.Create(testCtx, cIn)
	NoError(t, err)
	fmt.Printf("~~~ %+v\n", cAns)

	vIn = &VerifyQ{Id: cAns.Id, Password: cIn.Password}
	vAns, err = client.Verify(testCtx, vIn)
	NoError(t, err)
	fmt.Printf("~~~ %+v\n", vAns)
}
