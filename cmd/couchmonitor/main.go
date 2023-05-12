package main

import (
	"flag"
	"log"
	"net/http"

	// "regexp"
	"time"

	"cloudant.com/couchmonitor/internal/monitors"

	// Cloudant Go SDK
	"github.com/IBM/cloudant-go-sdk/cloudantv1"

	// Prometheus client
	"github.com/prometheus/client_golang/prometheus"
	// "github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var addr = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")

// entry point
func main() {
	log.Println("Hello, World!")

	// connect to Cloudant
	service, err := cloudantv1.NewCloudantV1UsingExternalConfig(
		&cloudantv1.CloudantV1Options{
			ServiceName: "CLOUDANT",
		})

	if err != nil {
		log.Fatalf("Could not initialise Cloudant client: %v", err)
	}

	log.Printf("Using Cloudant: %s", service.GetServiceURL())

	// Create a new registry.
	reg := prometheus.NewRegistry()

	// set up the replication collector to poll every 5s
	rc := monitors.ReplicationCollector{
		Reg:      reg,
		Cldt:     service,
		Interval: 5 * time.Second,
		Done:     make(chan bool),
	}
	rc.Go()

	// Expose the registered metrics via HTTP.
	http.Handle("/metrics", promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{
			// Opt into OpenMetrics to support exemplars.
			EnableOpenMetrics: true,
		},
	))
	log.Fatal(http.ListenAndServe(*addr, nil))
}
