package models

import (
	"context"
	// "fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Interceptor struct{}

func NewInterceptor() *Interceptor {
	return &Interceptor{}
}

func (inte *Interceptor) Unary() grpc.ServerOption {
	call := func(
		ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (
		resp any, err error) {

		var (
			ok    bool
			start time.Time
			md    metadata.MD
		)

		start = time.Now()

		if md, ok = metadata.FromIncomingContext(ctx); !ok {
			return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
		}

		log.Println(">>> Unary:", info.FullMethod, md)
		resp, err = handler(ctx, req)
		log.Println("<<<", time.Since(start), err)

		return resp, err
	}

	return grpc.UnaryInterceptor(call)
}

func (inte *Interceptor) Stream() grpc.ServerOption {
	call := func(
		srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler,
	) (err error) {

		var (
			ok    bool
			start time.Time
			md    metadata.MD
		)

		start = time.Now()

		if md, ok = metadata.FromIncomingContext(ss.Context()); !ok {
			return status.Errorf(codes.Unauthenticated, "authorization token is not provided")
		}

		log.Println(">>> Stream:", info.FullMethod, md)
		err = handler(srv, ss)
		log.Println("<<<", time.Since(start), err)

		return err
	}

	return grpc.StreamInterceptor(call)
}

func (inte *Interceptor) ClientUnary() grpc.DialOption {
	call := func(
		ctx context.Context, method string, req, reply any,
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption,
	) (err error) {

		log.Printf(">>> ClientUnaryUnary: %+v\n", method)
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "abcdefg")
		return invoker(ctx, method, req, reply, cc, opts...)
	}

	return grpc.WithUnaryInterceptor(call)
}

func (inte *Interceptor) ClientStream() grpc.DialOption {
	call := func(
		ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn,
		method string, streamer grpc.Streamer, opts ...grpc.CallOption,
	) (client grpc.ClientStream, err error) {

		log.Printf(">>> ClientStream: %+v\n", method)
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "abcdefg")
		return streamer(ctx, desc, cc, method, opts...)
	}

	return grpc.WithStreamInterceptor(call)
}
