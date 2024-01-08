package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/IBM/cloudant-go-sdk/cloudantv1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"cloudant.com/cloudant_exporter/internal/monitors"
)

var AppName = "cloudant_exporter"
var Version = "development"

var addr = flag.String("listen-address", "127.0.0.1:8080", "The address to listen on for HTTP requests.")

// entry point
func main() {
	log.Println(AppName)
	log.Printf("version %s(%s)", Version, runtime.Version())
	flag.Parse()

	cldt, err := newCloudantClient()
	if err != nil {
		log.Fatalf("Could not initialise Cloudant client: %v", err)
	}
	userAgent := fmt.Sprintf("%s/%s(%s)", AppName, Version, runtime.Version())
	cldt.Service.SetUserAgent(userAgent)

	log.Printf("Using Cloudant: %s", cldt.GetServiceURL())

	// Register all collectors ready to collect data
	rsc := monitors.NewReplicationStatusCollector(cldt)
	go func() { rsc.Start() }()
	prometheus.MustRegister(
		monitors.ReplicationProgressCollector{Cldt: cldt},
		monitors.ThroughputCollector{Cldt: cldt},
		monitors.ActiveTasksCollector{Cldt: cldt},
		rsc,
	)

	http.Handle("/metrics", promhttp.Handler())
	server := &http.Server{
		Addr:              *addr,
		ReadHeaderTimeout: 3 * time.Second,
	}
	log.Printf("HTTP server starting on %s", *addr)
	log.Fatal(server.ListenAndServe())
}

// newCloudantClient creates a new client for Cloudant, configured
// from environment variables, with a safe HTTP client.
func newCloudantClient() (*cloudantv1.CloudantV1, error) {

	// connect to Cloudant
	service, err := cloudantv1.NewCloudantV1UsingExternalConfig(
		&cloudantv1.CloudantV1Options{
			ServiceName: "CLOUDANT",
		},
	)
	if err != nil {
		return nil, err
	}

	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100
	t.MaxConnsPerHost = 10
	t.MaxIdleConnsPerHost = 10
	c := &http.Client{
		Timeout:   10 * time.Second,
		Transport: t,
	}
	service.Service.SetHTTPClient(c)

	service.EnableRetries(3, 30*time.Second)

	return service, nil
}
