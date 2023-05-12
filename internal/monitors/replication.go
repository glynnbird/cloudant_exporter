package monitors

import (
	"log"
	"time"

	"cloudant.com/couchmonitor/internal/utils"
	"github.com/IBM/cloudant-go-sdk/cloudantv1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type ReplicationMonitor struct {
	Cldt     *cloudantv1.CloudantV1
	Interval time.Duration
	FailBox  *utils.FailBox
}

var (
	docsProcessed = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cloudant_replication_docs_processed_total",
		Help: "The number of documents written to the target",
	},
		[]string{"docid"},
	)
)

func (rc *ReplicationMonitor) Go() {
	ticker := time.NewTicker(rc.Interval)

	for {
		select {
		case <-ticker.C:
			log.Println("ReplicationMonitor: tick")

			err := rc.tick()

			// Exit the monitor if we've not been successful for 20 minutes
			if err != nil {
				log.Printf("ReplicationMonitor error getting tasks: %v; last success: %s", err, rc.FailBox.LastSuccess())
				rc.FailBox.Failure()
				if rc.FailBox.ShouldExit() {
					log.Printf("ReplicationMonitor exiting; >20 minutes since last success at %s", rc.FailBox.LastSuccess())
					return
				}
			} else {
				rc.FailBox.Success()
			}
		}
	}
}

func (rc *ReplicationMonitor) tick() error {
	// fetch scheduler status
	getSchedulerDocsOptions := rc.Cldt.NewGetSchedulerDocsOptions()
	getSchedulerDocsOptions.SetLimit(50)
	getSchedulerDocsOptions.SetStates([]string{"running"})

	schedulerDocsResult, _, err := rc.Cldt.GetSchedulerDocs(getSchedulerDocsOptions)
	if err != nil {
		return err
	}
	for _, d := range schedulerDocsResult.Docs {
		log.Printf("ReplicationMonitor: Replication %q: docs written %d", *d.DocID, *d.Info.DocsWritten)
		docsProcessed.WithLabelValues(*d.DocID).Set(float64(*d.Info.DocsWritten))
	}
	return nil
}
