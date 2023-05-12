package monitors

import (
	"log"
	"time"

	"github.com/IBM/cloudant-go-sdk/cloudantv1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type ReplicationCollector struct {
	Cldt     *cloudantv1.CloudantV1
	Interval time.Duration
	Done     chan bool
}

var (
	docsProcessed = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cloudant_replication_docs_processed_total",
		Help: "The number of documents written to the target",
	},
		[]string{"docid"},
	)
)

func (rc *ReplicationCollector) Go() {
	ticker := time.NewTicker(rc.Interval)

	go func() {
		for {
			select {
			case <-rc.Done:
				return
			case t := <-ticker.C:
				log.Println("Polling Cloudant replication", t)

				// fetch scheduler status
				getSchedulerDocsOptions := rc.Cldt.NewGetSchedulerDocsOptions()
				getSchedulerDocsOptions.SetLimit(50)
				getSchedulerDocsOptions.SetStates([]string{"running"})

				schedulerDocsResult, _, err := rc.Cldt.GetSchedulerDocs(getSchedulerDocsOptions)
				if err != nil {
					log.Printf("ReplicationCollector: Error in GetSchedulerDocs: %v", err)
					continue
				}
				for _, d := range schedulerDocsResult.Docs {
					log.Printf("Replication %q: docs written %d", *d.DocID, *d.Info.DocsWritten)
					docsProcessed.WithLabelValues(*d.DocID).Set(float64(*d.Info.DocsWritten))
				}
			}
		}
	}()

}
