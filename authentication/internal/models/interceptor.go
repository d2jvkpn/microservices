package models

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Interceptor struct {
	logger *zap.Logger
}

func NewInterceptor(logger *zap.Logger) *Interceptor {
	return &Interceptor{logger}
}

func (inte *Interceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (
		resp any, err error) {

		var (
			ok    bool
			start time.Time
			// md    metadata.MD
		)

		start = time.Now()

		if _, ok = metadata.FromIncomingContext(ctx); !ok {
			return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
		}

		resp, err = handler(ctx, req)
		if err == nil {
			inte.logger.Info(info.FullMethod, zap.Duration("latency", time.Since(start)))
		} else {
			inte.logger.Error(
				info.FullMethod, zap.Duration("latency", time.Since(start)), zap.Any("error", err),
			)
		}

		return resp, err
	}

	// return grpc.UnaryInterceptor(call) // grpc.ServerOption
}

func (inte *Interceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler,
	) (err error) {

		var (
			ok    bool
			start time.Time
			// md    metadata.MD
		)

		start = time.Now()

		if _, ok = metadata.FromIncomingContext(ss.Context()); !ok {
			return status.Errorf(codes.Unauthenticated, "authorization token is not provided")
		}

		err = handler(srv, ss)
		latency := fmt.Sprintf("%s", time.Since(start))

		if err == nil {
			inte.logger.Info(info.FullMethod, zap.String("latency", latency))
		} else {
			inte.logger.Error(
				info.FullMethod, zap.String("latency", latency), zap.Any("error", err),
			)
		}

		return err
	}

	// return grpc.StreamInterceptor(call) // grpc.ServerOption
}
