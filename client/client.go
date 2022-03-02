package client

import (
	"context"
	"fmt"
	"time"

	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/newrelic/go-agent/v3/integrations/nrgrpc"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultTimeout = 10 * time.Second
	defaultRetries = 5
)

// Client is a wrapper over grpc.Client
// it provides a set of function for redialing, backoff, and adding a set of
// interceptors (unary and stream) to the grpc.Client.
// Example usage:
// client := grpcclient.New(true).AddUnaryChains(someUnaryInterceptors_here...).SetMaxRetries(3)
// logger := logrus.New().WithField("app", "test")
// if err := client.Connect(SOME_SERVICE_ADDRESS, logger); err != nil {
// 		return err
// }
type Client struct {
	*grpc.ClientConn
	unaryChains  []grpc.UnaryClientInterceptor
	streamChains []grpc.StreamClientInterceptor
	timeout      time.Duration
	maxRetries   uint
}

// New creates a new instance of grpc.Client, it takes a single argument
// which is whether or not should we enable newrelic interceptor by default.
func New(newrelicEnabled bool) *Client {
	c := &Client{
		timeout:    defaultTimeout,
		maxRetries: defaultRetries,
	}
	if newrelicEnabled {
		c.AddUnaryChains(nrgrpc.UnaryClientInterceptor)
		c.AddStreamChains(nrgrpc.StreamClientInterceptor)
	}
	return c
}

// AddUnaryChains is a builder method returning Client instance with unary middlewares added
func (c *Client) AddUnaryChains(chains ...grpc.UnaryClientInterceptor) *Client {
	c.unaryChains = append(c.unaryChains, chains...)
	return c
}

// AddStreamChains is a builder method returning Client instance with stream middlewares added
func (c *Client) AddStreamChains(chains ...grpc.StreamClientInterceptor) *Client {
	c.streamChains = append(c.streamChains, chains...)
	return c
}

// SetMaxRetries sets maximum number of retries a grpc.Client should do
func (c *Client) SetMaxRetries(maxRetries uint) *Client {
	c.maxRetries = maxRetries
	return c
}

// Connect creates a new client connection, resulting client will be returned alongside error if exists
func (c *Client) Connect(serverAddr string, log *logrus.Entry) error {
	// Adding logrus as default
	grpc_logrus.ReplaceGrpcLogger(log)
	c.AddUnaryChains(grpc_logrus.UnaryClientInterceptor(log))

	c.expBackOff()

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()
	conn, err := grpc.DialContext(
		ctx,
		serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(c.unaryChains...),
		grpc.WithChainStreamInterceptor(c.streamChains...),
	)
	if err != nil {
		return fmt.Errorf("grpc.DialContext(): %w", err)
	}
	c.ClientConn = conn
	return nil
}

func (c *Client) expBackOff() {
	opts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(500 * time.Millisecond)),
		grpc_retry.WithMax(c.maxRetries),
	}
	c.AddUnaryChains(grpc_retry.UnaryClientInterceptor(opts...))
	c.AddStreamChains(grpc_retry.StreamClientInterceptor(opts...))
}
