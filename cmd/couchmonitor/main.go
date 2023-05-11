package main

import (
	"encoding/json"
	"flag"
	"fmt"
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

// poll the Cloudant replication scheduler every 5 seconds
func Collect(service *cloudantv1.CloudantV1) {

	// nicked from https://gobyexample.com/tickers
	ticker := time.NewTicker(5000 * time.Millisecond)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				fmt.Println("Polling Cloudant replication", t)

				// fetch scheduler status
				getSchedulerDocsOptions := service.NewGetSchedulerDocsOptions()
				schedulerDocsResult, _, _ := service.GetSchedulerDocs(getSchedulerDocsOptions)
				b, _ := json.MarshalIndent(schedulerDocsResult, "", "  ")

				// to stdout - not plumbed into Prometheus client yet
				fmt.Println(string(b))
			}
		}
	}()
}

// entry point
func main() {
	log.Println("Hello, World!")

	// connect to Cloudant
	service, _ := cloudantv1.NewCloudantV1UsingExternalConfig(
		&cloudantv1.CloudantV1Options{
			ServiceName: "CLOUDANT",
		})

	// collectors - pass in the Cloudant service

	Collect(service)

	// Create a new registry.
	reg := prometheus.NewRegistry()

	rc := monitors.ReplicationCollector{
		Reg:      reg,
		Cldt:     service,
		Interval: 5 * time.Second,
		Done:     make(chan bool),
	}
	rc.Go()

	// Add Go module build info.
	// reg.MustRegister(collectors.NewBuildInfoCollector())
	// reg.MustRegister(collectors.NewGoCollector(
	// 	collectors.WithGoCollectorRuntimeMetrics(collectors.GoRuntimeMetricsRule{Matcher: regexp.MustCompile("/.*")}),
	// ))

	// Expose the registered metrics via HTTP.
	http.Handle("/metrics", promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{
			// Opt into OpenMetrics to support exemplars.
			EnableOpenMetrics: true,
		},
	))
	log.Println("Hello world from new Go Collector!")
	log.Fatal(http.ListenAndServe(*addr, nil))
}
