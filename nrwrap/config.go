package nrwrap

import (
	"errors"

	"github.com/newrelic/go-agent/v3/integrations/nrlogrus"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/sirupsen/logrus"
)

// Config holds the newrelic config
type Config struct {
	AppName   string
	SecretKey string
}

// TxProducer is just a small wrapper around *newrelic.Application, as we don't want non-http / grpc services
// to encapsulate transaction producer
type TxProducer interface {
	StartTransaction(string) *newrelic.Transaction
}

// InitApplication initializes the newrelic for service tracing
func InitApplication(cfg Config, logs *logrus.Logger) (*newrelic.Application, error) {
	if cfg.AppName == "" {
		return nil, errors.New("new relic app name is empty")
	}

	if cfg.SecretKey == "" {
		return nil, errors.New("new relic secret key is empty")
	}

	var err error
	app, err := newrelic.NewApplication(
		newrelic.ConfigEnabled(true),
		newrelic.ConfigAppName(cfg.AppName),
		newrelic.ConfigLicense(cfg.SecretKey),
		newrelic.ConfigDistributedTracerEnabled(true),
		func(config *newrelic.Config) {
			config.Enabled = true
			config.Logger = nrlogrus.Transform(logs)
			config.ErrorCollector.Enabled = true
			config.ErrorCollector.CaptureEvents = true
			config.ErrorCollector.RecordPanics = true
			config.TransactionTracer.Enabled = true
		},
	)
	if err != nil {
		return nil, err
	}
	return app, nil
}
