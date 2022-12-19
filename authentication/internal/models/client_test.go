package models

import (
	"fmt"
	"testing"

	. "authentication/proto"

	. "github.com/stretchr/testify/require"
	"google.golang.org/grpc"
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
		guq  *GetOrUpdateQ
		gua  *GetOrUpdateA
	)

	inte := NewInterceptor()

	conn, err = grpc.Dial(testAddr,
		grpc.WithInsecure(),
		inte.ClientUnary(),
		inte.ClientStream(),
	)
	NoError(t, err)

	client = NewAuthServiceClient(conn)
	cIn = &CreateQ{Password: "123456"}

	cAns, err = client.Create(testCtx, cIn)
	NoError(t, err)
	fmt.Printf("~~~ Create: %+v\n", cAns)

	vIn = &VerifyQ{Id: cAns.Id, Password: cIn.Password}
	vAns, err = client.Verify(testCtx, vIn)
	NoError(t, err)
	fmt.Printf("~~~ Verify: %+v\n", vAns)

	guq = &GetOrUpdateQ{Id: cAns.Id}
	gua, err = client.GetOrUpdate(testCtx, guq)
	NoError(t, err)
	fmt.Printf("~~~ GetOrUpdate: %+v\n", gua)

	guq = &GetOrUpdateQ{Id: cAns.Id, Status: "blocked"}
	_, err = client.GetOrUpdate(testCtx, guq)
	NoError(t, err)
}
