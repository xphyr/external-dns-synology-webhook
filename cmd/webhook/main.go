package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/codingconcepts/env"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/xphyr/external-dns-synology-webhook/internal/server"
	"github.com/xphyr/external-dns-synology-webhook/internal/synology"
	"sigs.k8s.io/external-dns/provider/webhook/api"
)

var logLevel string

func init() {
	flag.StringVar(&logLevel, "loglevel", "info", "Set the log level (trace, debug, info, warn, error, fatal, panic)")
}

// loop waits for a SIGTERM or a SIGINT and then shuts down the server.
func loop(status *server.HealthStatus) {
	exitSignal := make(chan os.Signal, 1)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	signal := <-exitSignal

	log.Infof("Signal %s received. Shutting down the webhook.", signal.String())
	status.SetHealth(false)
	status.SetReady(false)
}

func main() {

	flag.Parse() // Parse the command-line flags

	// Parse the log level string into a logrus.Level
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		fmt.Printf("Invalid log level specified: %s. Defaulting to info.\n", logLevel)
		logrus.SetLevel(logrus.InfoLevel) // Set default if parsing fails
	} else {
		logrus.SetLevel(level)
	}

	// Read server options
	serverOptions := &server.ServerOptions{}

	if err := env.Set(serverOptions); err != nil {
		log.Fatal(err)
	}

	// Start health server
	log.Infof("Starting liveness and readiness server on %s", serverOptions.GetHealthAddress())
	healthStatus := server.HealthStatus{}
	healthServer := server.HealthServer{}
	go healthServer.Start(&healthStatus, nil, *serverOptions)

	// Read provider configuration
	providerConfig := &synology.Configuration{}
	if err := env.Set(providerConfig); err != nil {
		log.Fatal(err)
	}

	// instantiate the synology provider
	provider := synology.NewProvider(providerConfig)

	// Start the webhook
	log.Infof("Starting webhook server on %s", serverOptions.GetWebhookAddress())
	startedChan := make(chan struct{})
	go api.StartHTTPApi(
		provider, startedChan,
		serverOptions.GetReadTimeout(),
		serverOptions.GetWriteTimeout(),
		serverOptions.GetWebhookAddress(),
	)

	// Wait for the HTTP server to start and then set the healthy and ready flags
	<-startedChan
	healthStatus.SetHealth(true)
	healthStatus.SetReady(true)

	// Loops until a signal tells us to exit
	loop(&healthStatus)
}
