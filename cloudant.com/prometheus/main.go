package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	// this is wrong
	//"collectors/replication"

	// Cloudant Go SDK
	"github.com/IBM/cloudant-go-sdk/cloudantv1"
)

func main() {
	fmt.Println("Hello, World!")

	service, _ := cloudantv1.NewCloudantV1UsingExternalConfig(
		&cloudantv1.CloudantV1Options{
			ServiceName: "CLOUDANT",
		})

	getSchedulerDocsOptions := service.NewGetSchedulerDocsOptions()
	schedulerDocsResult, _, _ := service.GetSchedulerDocs(getSchedulerDocsOptions)
	b, _ := json.MarshalIndent(schedulerDocsResult, "", "  ")
	//fmt.Println(string(b))

	// collectors
	// this doesn't work
	//replication()

	// web server to handle Prometheus GET /metrics endpoint
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, string(b))
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
