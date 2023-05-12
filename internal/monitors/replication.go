package monitors

import (
	"log"
	"time"

	"github.com/IBM/cloudant-go-sdk/cloudantv1"
	"github.com/prometheus/client_golang/prometheus"
)

type ReplicationCollector struct {
	Reg      *prometheus.Registry
	Cldt     *cloudantv1.CloudantV1
	Interval time.Duration
	Done     chan bool
}

func (rc *ReplicationCollector) Go() {
	// dummy labels (for now)
	labels := make(map[string]string)
	labels["id"] = "dallas_to_washington_prod"

	docsProcessed := prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "cloudant_replication_docs_processed_total",
		Help:        "The number of documents written to the target",
		ConstLabels: prometheus.Labels(labels),
	})
	rc.Reg.MustRegister(docsProcessed)

	ticker := time.NewTicker(rc.Interval)

	go func() {
		for {
			select {
			case <-rc.Done:
				return
			case t := <-ticker.C:
				log.Println("Tick at", t)
				log.Println("Polling Cloudant replication", t)

				// fetch scheduler status
				getSchedulerDocsOptions := rc.Cldt.NewGetSchedulerDocsOptions()
				getSchedulerDocsOptions.SetLimit(50)
				getSchedulerDocsOptions.SetStates([]string{"running"})
				schedulerDocsResult, _, err := rc.Cldt.GetSchedulerDocs(getSchedulerDocsOptions)

				if err != nil {
					log.Printf("Error in GetSchedulerDocs: %v", err)
					continue
				}

				if len(schedulerDocsResult.Docs) > 0 {
					log.Printf("docs written %d", *schedulerDocsResult.Docs[0].Info.DocsWritten)
					docsProcessed.Set(float64(*schedulerDocsResult.Docs[0].Info.DocsWritten))
				}
			}
		}
	}()

}
