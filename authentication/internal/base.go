package internal

import (
	// "fmt"

	"github.com/d2jvkpn/go-web/pkg/wrap"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"google.golang.org/grpc"
)

var (
	_Logger       *zap.Logger
	_DB           *gorm.DB
	_ConsulClient *wrap.ConsulClient
	_CloseTracer  func()
	_GrpcServer   *grpc.Server
)
