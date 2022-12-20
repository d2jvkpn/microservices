package models

import (
	"context"
	"fmt"
	"testing"
	"time"

	"authentication/internal/settings"
	. "authentication/proto"

	"github.com/d2jvkpn/go-web/pkg/cloud_native"
	"github.com/spf13/viper"
	. "github.com/stretchr/testify/require"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

func TestClient(t *testing.T) {
	var (
		enableOtel  bool
		closeTracer func()
		err         error
		conn        *grpc.ClientConn
		client      AuthServiceClient

		cIn  *CreateQ
		cAns *CreateA
		vIn  *VerifyQ
		vAns *VerifyA
		guq  *GetOrUpdateQ
		gua  *GetOrUpdateA
	)

	enableOtel = testConfig.GetBool("opentelemetry.enable")
	if enableOtel {
		if closeTracer, err = testLoadOtel(testConfig); err != nil {
			return
		}
		defer closeTracer()
	}

	inte := NewClientInterceptor()

	conn, err = grpc.Dial(testAddr,
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
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

func testLoadOtel(vc *viper.Viper) (closeTracer func(), err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	str := vc.GetString("opentelemetry.address")
	secure := vc.GetBool("opentelemetry.secure")

	closeTracer, err = cloud_native.LoadTracer(ctx, str, settings.App, secure)
	if err != nil {
		return nil, fmt.Errorf("cloud_native.LoadTracer: %s, %w", str, err)
	}

	return closeTracer, nil
}
