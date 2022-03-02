package server

import (
	"context"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/newrelic/go-agent/v3/integrations/nrgrpc"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

const (
	defaultGracefulShutdownTimeout = 10 * time.Second
)

// Server is a wrapper around generic grpc.Server.
type Server struct {
	grpcSrv      *grpc.Server
	log          *logrus.Entry
	unaryChains  []grpc.UnaryServerInterceptor
	streamChains []grpc.StreamServerInterceptor

	Agent                   *newrelic.Application // we provide Agent as exportable field here
	LogPayload              bool
	GracefulShutdownTimeout time.Duration
}

func New(log *logrus.Entry) *Server {
	return &Server{
		log:                     log,
		GracefulShutdownTimeout: defaultGracefulShutdownTimeout,
	}
}

func (s *Server) AddUnaryChains(chains ...grpc.UnaryServerInterceptor) {
	s.unaryChains = append(s.unaryChains, chains...)
}

func (s *Server) AddStreamChains(chains ...grpc.StreamServerInterceptor) {
	s.streamChains = append(s.streamChains, chains...)

}

func (s *Server) Create() (*grpc.Server, error) {
	// Log gRPC library internals with logrus
	grpc_logrus.ReplaceGrpcLogger(s.log)

	s.AddUnaryChains(grpc_ctxtags.UnaryServerInterceptor(
		grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		grpc_logrus.UnaryServerInterceptor(s.log),
		otelgrpc.UnaryServerInterceptor(),
		grpc_recovery.UnaryServerInterceptor(),
	)

	// inject new relic middleware
	if s.Agent != nil {
		s.AddUnaryChains(nrgrpc.UnaryServerInterceptor(s.Agent))
		s.AddStreamChains(nrgrpc.StreamServerInterceptor(s.Agent))
	}

	if s.LogPayload {
		decider := func(ctx context.Context, fullMethodName string, servingObject interface{}) bool {
			return true
		}

		s.AddUnaryChains(grpc_logrus.PayloadUnaryServerInterceptor(s.log, decider))
	}

	opts := []grpc.ServerOption{
		grpc_middleware.WithUnaryServerChain(s.unaryChains...),
		grpc_middleware.WithStreamServerChain(s.streamChains...),
	}

	s.grpcSrv = grpc.NewServer(opts...)
	return s.grpcSrv, nil
}

func (s *Server) Shutdown() bool {
	c := make(chan struct{})

	go func() {
		defer close(c)

		// Block until all pending RPCs are finished
		s.grpcSrv.GracefulStop()
	}()

	select {
	case <-time.After(s.GracefulShutdownTimeout):
		// Timeout
		s.grpcSrv.Stop()
		<-c
		return false

	case <-c:
		// Shutdown completed within the timeout
		return true
	}
}
