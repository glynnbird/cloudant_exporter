package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/IBM/cloudant-go-sdk/cloudantv1"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"cloudant.com/couchmonitor/internal/monitors"
	"cloudant.com/couchmonitor/internal/utils"
)

var addr = flag.String("listen-address", "127.0.0.1:8080", "The address to listen on for HTTP requests.")

const failAfter = 5 * time.Minute

// entry point
func main() {
	log.Println("Hello, World!")

	cldt, err := newCloudantClient()
	if err != nil {
		log.Fatalf("Could not initialise Cloudant client: %v", err)
	}

	log.Printf("Using Cloudant: %s", cldt.GetServiceURL())

	// Monitors publish to this channel if they fail,
	// typically that they haven't made a successful
	// request in `failAfter` time.
	monitorFailed := make(chan string)

	rc := monitors.ReplicationMonitor{
		Cldt:     cldt,
		Interval: 5 * time.Second,
		FailBox:  utils.NewFailBox(failAfter),
	}
	go func() {
		rc.Go()
		monitorFailed <- "ReplicationMonitor"
	}()

	tm := monitors.ThroughputMonitor{
		Cldt:     cldt,
		Interval: 5 * time.Second,
		FailBox:  utils.NewFailBox(failAfter),
	}
	go func() {
		tm.Go()
		monitorFailed <- "ThroughputMonitor"
	}()

	atm := monitors.ActiveTasksMonitor{
		Cldt:     cldt,
		Interval: 5 * time.Second,
		FailBox:  utils.NewFailBox(failAfter),
	}
	go func() {
		atm.Go()
		monitorFailed <- "ActiveTasksMonitor"
	}()

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		log.Fatal(http.ListenAndServe(*addr, nil))
	}()
	log.Printf("HTTP server started on %s", *addr)

	// After a monitor fails, we need to shutdown.
	m := <-monitorFailed
	log.Printf("A monitor died: %q! Exiting.", m)
	// exiting main kills everything
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
