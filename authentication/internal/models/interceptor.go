package models

import (
	"context"
	"fmt"
	"time"

	"github.com/d2jvkpn/microservices/authentication/internal/settings"

	"go.uber.org/zap"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Interceptor struct{}

func NewInterceptor() *Interceptor {
	return &Interceptor{}
}

func (inte *Interceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
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

		logger := settings.Logger.Named("interceptor")
		logger.Debug("models.Interceptor.Unary", zap.Any("md", md))

		resp, err = handler(ctx, req)
		latency := fmt.Sprintf("%s", time.Since(start))

		if err == nil {
			logger.Info(info.FullMethod, zap.String("latency", latency))
		} else {
			logger.Error(
				info.FullMethod, zap.String("latency", latency), zap.Any("error", err),
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
			md    metadata.MD
		)

		start = time.Now()

		if md, ok = metadata.FromIncomingContext(ss.Context()); !ok {
			return status.Errorf(codes.Unauthenticated, "authorization token is not provided")
		}

		logger := settings.Logger.Named("interceptor")
		logger.Debug("models.Interceptor.Stream", zap.Any("md", md))

		err = handler(srv, ss)
		latency := fmt.Sprintf("%s", time.Since(start))

		if err == nil {
			logger.Info(info.FullMethod, zap.String("latency", latency))
		} else {
			logger.Error(
				info.FullMethod, zap.String("latency", latency), zap.Any("error", err),
			)
		}

		return err
	}

	// return grpc.StreamInterceptor(call) // grpc.ServerOption
}
