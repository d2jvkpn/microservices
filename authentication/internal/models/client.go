package models

import (
	"context"
	// "fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ClientInterceptor struct{}

func NewClientInterceptor() *ClientInterceptor {
	return &ClientInterceptor{}
}

func (inte *ClientInterceptor) ClientUnary(kv ...string) grpc.DialOption {
	call := func(
		ctx context.Context, method string, req, reply any,
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption,
	) (err error) {

		ctx = metadata.AppendToOutgoingContext(ctx, kv...)
		err = invoker(ctx, method, req, reply, cc, opts...)

		return err
	}

	return grpc.WithUnaryInterceptor(call)
}

func (inte *ClientInterceptor) ClientStream(kv ...string) grpc.DialOption {
	call := func(
		ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn,
		method string, streamer grpc.Streamer, opts ...grpc.CallOption,
	) (client grpc.ClientStream, err error) {

		ctx = metadata.AppendToOutgoingContext(ctx, kv...)
		client, err = streamer(ctx, desc, cc, method, opts...)

		return client, err
	}

	return grpc.WithStreamInterceptor(call)
}
