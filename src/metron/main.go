package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"metric"
	"time"

	"google.golang.org/grpc"

	"metron/api"
	"plumbing"
	"profiler"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	configFilePath := flag.String(
		"config",
		"config/metron.json",
		"Location of the Metron config json file",
	)
	flag.Parse()

	config, err := api.ParseConfig(*configFilePath)
	if err != nil {
		log.Fatalf("Unable to parse config: %s", err)
	}

	clientCreds := plumbing.NewCredentials(
		config.GRPC.CertFile,
		config.GRPC.KeyFile,
		config.GRPC.CAFile,
		"doppler",
	)
	serverCreds := plumbing.NewCredentials(
		config.GRPC.CertFile,
		config.GRPC.KeyFile,
		config.GRPC.CAFile,
		"metron",
	)

	appV1 := api.NewV1App(config, clientCreds)
	go appV1.Start()

	appV2 := api.NewV2App(config, clientCreds, serverCreds)
	go appV2.Start()

	batchInterval := time.Duration(config.MetricBatchIntervalMilliseconds) * time.Millisecond
	metric.Setup(
		metric.WithGrpcDialOpts(grpc.WithTransportCredentials(serverCreds)),
		metric.WithBatchInterval(batchInterval),
		metric.WithPrefix("loggregator"),
		metric.WithOrigin("metron"),
		metric.WithAddr(fmt.Sprintf("localhost:%d", config.GRPC.Port)),
		metric.WithDeploymentMeta(config.Deployment, config.Job, config.Index),
	)

	// We start the profiler last so that we can definitively say that we're
	// all connected and ready for data by the time the profiler starts up.
	profiler.New(config.PPROFPort).Start()
}
