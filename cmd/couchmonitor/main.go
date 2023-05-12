package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/IBM/cloudant-go-sdk/cloudantv1"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"cloudant.com/couchmonitor/internal/monitors"
)

var addr = flag.String("listen-address", "127.0.0.1:8080", "The address to listen on for HTTP requests.")

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

	// set up the replication collector to poll every 5s
	rc := monitors.ReplicationCollector{
		Cldt:     service,
		Interval: 5 * time.Second,
		Done:     make(chan bool),
	}
	rc.Go()
	tm := monitors.ThroughputMonitor{
		Cldt:     service,
		Interval: 5 * time.Second,
		Done:     make(chan bool),
	}
	tm.Go()

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))
}
