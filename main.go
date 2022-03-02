package main

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"

	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/newrelic/go-agent/v3/integrations/logcontext/nrlogrusplugin"
	"github.com/newrelic/go-agent/v3/integrations/nrgrpc/example/sampleapp"
	"github.com/newrelic/go-agent/v3/newrelic"
	log "github.com/sirupsen/logrus"

	"github.com/alexadhy/rf-arch-example-test/client"
	"github.com/alexadhy/rf-arch-example-test/nrwrap"
	"github.com/alexadhy/rf-arch-example-test/server"
)

const (
	svcName = "rf-arch-example-test"
)

// processMessage processes each incoming Message.
func processMessage(ctx context.Context, msg *sampleapp.Message, logger *log.Logger, i int) {
	txn := newrelic.FromContext(ctx)
	txn.StartSegment("processMessage")
	c := newrelic.NewContext(context.Background(), txn)
	if i != 1 {
		logger.WithContext(c).Infof("Message received: %s\n", msg.Text)
	} else {
		txn.NoticeError(fmt.Errorf("simulated error"))
		logger.WithContext(c).Error(fmt.Errorf("simulated error"))
	}
	txn.End()
}

// Server is a gRPC server.
type Server struct {
	logger *log.Logger
}

// DoUnaryUnary is a unary request, unary response method.
func (s *Server) DoUnaryUnary(ctx context.Context, msg *sampleapp.Message) (*sampleapp.Message, error) {

	processMessage(ctx, msg, s.logger, rand.Intn(4))
	return &sampleapp.Message{Text: "Hello from DoUnaryUnary"}, nil
}

// DoUnaryStream is a unary request, stream response method.
func (s *Server) DoUnaryStream(msg *sampleapp.Message, stream sampleapp.SampleApplication_DoUnaryStreamServer) error {
	return nil
}

// DoStreamUnary is a stream request, unary response method.
func (s *Server) DoStreamUnary(stream sampleapp.SampleApplication_DoStreamUnaryServer) error {
	// TODO
	return nil
}

// DoStreamStream is a stream request, stream response method.
func (s *Server) DoStreamStream(stream sampleapp.SampleApplication_DoStreamStreamServer) error {
	return nil
}

func mustGetEnv(key string) string {
	if val := os.Getenv(key); "" != val {
		return val
	}
	panic(fmt.Sprintf("environment variable %s unset", key))
}

func main() {
	cfg := nrwrap.Config{
		AppName:   svcName,
		SecretKey: mustGetEnv("NEW_RELIC_LICENSE_KEY"),
	}
	logger := log.New()
	logger.SetFormatter(nrlogrusplugin.ContextFormatter{})
	app, err := nrwrap.InitApplication(cfg, logger)
	if err != nil {
		panic(err)
	}

	// Log gRPC library internals with logrus
	entry := logger.WithField("service", svcName)
	grpc_logrus.ReplaceGrpcLogger(entry)

	lis, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}

	g := server.New(entry)
	g.Agent = app

	grpcServer, err := g.Create()
	if err != nil {
		panic(err)
	}

	sampleapp.RegisterSampleApplicationServer(grpcServer, &Server{
		logger: logger,
	})

	tc := time.NewTicker(10 * time.Second)
	doneCh := make(chan struct{}, 1)

	go func() {
		for {
			select {
			case <-doneCh:
				tc.Stop()
				return
			case <-tc.C:
				startClient(entry)
			}
		}
	}()

	if err = grpcServer.Serve(lis); err != nil {
		log.Println(err)
	}

}

func doUnaryUnary(ctx context.Context, client sampleapp.SampleApplicationClient, logger *log.Logger) {
	msg, err := client.DoUnaryUnary(ctx, &sampleapp.Message{Text: "Hello DoUnaryUnary"})
	if nil != err {
		logger.WithContext(ctx).WithField("type", "client").Info(err)
		return
	}
	logger.WithContext(ctx).WithField("type", "client").Info(msg.Text)
}

func startClient(logger *log.Entry) {
	conn := client.New(true)
	if err := conn.Connect("localhost:8080", logger); err != nil {
		logger.WithError(err).Error("grpc.Dial() : %v", err)
		return
	}

	cfg := nrwrap.Config{
		AppName:   "rf-arch-example-test-client",
		SecretKey: mustGetEnv("NEW_RELIC_LICENSE_KEY"),
	}
	app, err := nrwrap.InitApplication(cfg, logger.Logger)
	if nil != err {
		logger.Println(err)
		return
	}
	err = app.WaitForConnection(30 * time.Second)
	if nil != err {
		logger.Println(err)
		return
	}
	defer func() {
		app.Shutdown(30 * time.Second)
		_ = conn.Close()
	}()

	txn := app.StartTransaction("client")
	appClient := sampleapp.NewSampleApplicationClient(conn.ClientConn)
	ctx := newrelic.NewContext(context.Background(), txn)
	doUnaryUnary(ctx, appClient, logger.Logger)
	txn.End()
}
