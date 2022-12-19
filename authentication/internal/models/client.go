package models

import (
	"context"
	// "fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ClientInterceptor struct{}

func NewClientInterceptor() *ClientInterceptor {
	return &ClientInterceptor{}
}

func (inte *ClientInterceptor) ClientUnary() grpc.DialOption {
	call := func(
		ctx context.Context, method string, req, reply any,
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption,
	) (err error) {

		start := time.Now()
		log.Printf(">>> ClientUnaryUnary: %+v\n", method)
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "abcdefg")

		err = invoker(ctx, method, req, reply, cc, opts...)
		log.Printf("<<< %s, %v\n", time.Since(start), err)

		return err
	}

	return grpc.WithUnaryInterceptor(call)
}

func (inte *ClientInterceptor) ClientStream() grpc.DialOption {
	call := func(
		ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn,
		method string, streamer grpc.Streamer, opts ...grpc.CallOption,
	) (client grpc.ClientStream, err error) {

		start := time.Now()
		log.Printf(">>> ClientStream: %+v\n", method)
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "abcdefg")
		client, err = streamer(ctx, desc, cc, method, opts...)
		log.Printf("<<< %s, %v\n", time.Since(start), err)

		return client, err
	}

	return grpc.WithStreamInterceptor(call)
}
