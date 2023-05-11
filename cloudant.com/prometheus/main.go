package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	// this is wrong
	//"collectors/replication"

	// Cloudant Go SDK
	"time"

	"github.com/IBM/cloudant-go-sdk/cloudantv1"
)

// poll the Cloudant replication scheduler every 5 seconds
func Collect(service *cloudantv1.CloudantV1) {

	ticker := time.NewTicker(5000 * time.Millisecond)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				fmt.Println("Polling Cloudant replication", t)

				getSchedulerDocsOptions := service.NewGetSchedulerDocsOptions()
				schedulerDocsResult, _, _ := service.GetSchedulerDocs(getSchedulerDocsOptions)
				b, _ := json.MarshalIndent(schedulerDocsResult, "", "  ")
				fmt.Println(string(b))
			}
		}
	}()
}

// entry point
func main() {
	fmt.Println("Hello, World!")

	// connect to Cloudant
	service, _ := cloudantv1.NewCloudantV1UsingExternalConfig(
		&cloudantv1.CloudantV1Options{
			ServiceName: "CLOUDANT",
		})

	// collectors
	Collect(service)

	// web server to handle Prometheus GET /metrics endpoint
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "metrics!")
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
