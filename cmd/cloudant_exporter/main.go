package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/IBM/cloudant-go-sdk/cloudantv1"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"cloudant.com/cloudant_exporter/internal/monitors"
	"cloudant.com/cloudant_exporter/internal/utils"
)

var Version = "development"

var addr = flag.String("listen-address", "127.0.0.1:8080", "The address to listen on for HTTP requests.")

const failAfter = 5 * time.Minute

// entry point
func main() {
	log.Println("Cloudant Prometheus Exporter")
	log.Printf("version %s", Version)
	flag.Parse()

	cldt, err := newCloudantClient()
	if err != nil {
		log.Fatalf("Could not initialise Cloudant client: %v", err)
	}
	cldt.Service.SetUserAgent(fmt.Sprintf("cloudant_exporter/%s", Version))

	log.Printf("Using Cloudant: %s", cldt.GetServiceURL())

	// Monitors publish to this channel if they fail,
	// typically that they haven't made a successful
	// request in `failAfter` time.
	monitorFailed := make(chan string)

	rc := monitorLooper{
		Interval: 5 * time.Second,
		FailBox:  utils.NewFailBox(failAfter),
		Chk:      &monitors.ReplicationMonitor{Cldt: cldt},
	}
	go func() {
		rc.Go()
		monitorFailed <- "ReplicationMonitor"
	}()

	tm := monitorLooper{
		Interval: 5 * time.Second,
		FailBox:  utils.NewFailBox(failAfter),
		Chk:      &monitors.ThroughputMonitor{Cldt: cldt},
	}
	go func() {
		tm.Go()
		monitorFailed <- "ThroughputMonitor"
	}()

	atm := monitorLooper{
		Interval: 5 * time.Second,
		FailBox:  utils.NewFailBox(failAfter),
		Chk:      &monitors.ActiveTasksMonitor{Cldt: cldt},
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

type monitor interface {
	Retrieve() error
	Name() string
}

// monitorLooper runs Chk every Interval, using FailBox to decide when to give up and exit
// on receiving errors.
type monitorLooper struct {
	Interval time.Duration
	FailBox  *utils.FailBox
	Chk      monitor
}

func (rc *monitorLooper) Go() {
	ticker := time.NewTicker(rc.Interval)

	for {
		select {
		case <-ticker.C:
			log.Printf("[%s] tick", rc.Chk.Name())
			err := rc.Chk.Retrieve()

			// Exit the monitor if we've not been successful for 20 minutes
			if err != nil {
				log.Printf("[%s] error getting tasks: %v; last success: %s", rc.Chk.Name(), err, rc.FailBox.LastSuccess())
				rc.FailBox.Failure()
			} else {
				rc.FailBox.Success()
			}

			if rc.FailBox.ShouldExit() {
				log.Printf("[%s] exiting; >%s since last success at %s", rc.Chk.Name(), failAfter, rc.FailBox.LastSuccess())
				return
			}
		}
	}
}
