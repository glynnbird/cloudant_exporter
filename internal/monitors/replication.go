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
	changesPendingTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cloudant_replication_changes_pending_total",
		Help: "The number of changes remaining to process (approximately)",
	},
		[]string{"docid"},
	)
	docWriteFailuresTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cloudant_replication_doc_write_failures_total",
		Help: "The number of failures writing documents to the target",
	},
		[]string{"docid"},
	)
	docsReadTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cloudant_replication_docs_read_total",
		Help: "Total number of documents read from the source database",
	},
		[]string{"docid"},
	)
	docsWrittenTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cloudant_replication_docs_written_total",
		Help: "Total number of documents written to the target database",
	},
		[]string{"docid"},
	)
	missingRevsFoundTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cloudant_replication_missing_revs_found_total",
		Help: "Total number of revs found so far on the source that are not at the target",
	},
		[]string{"docid"},
	)
	revsCheckedTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cloudant_replication_revs_checked_total",
		Help: "Total number of revs processed on the source",
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
			} else {
				rc.FailBox.Success()
			}

			if rc.FailBox.ShouldExit() {
				log.Printf("ReplicationMonitor exiting; >20 minutes since last success at %s", rc.FailBox.LastSuccess())
				return
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
		changesPendingTotal.WithLabelValues(*d.DocID).Set(float64(*d.Info.ChangesPending))
		docWriteFailuresTotal.WithLabelValues(*d.DocID).Set(float64(*d.Info.DocWriteFailures))
		docsReadTotal.WithLabelValues(*d.DocID).Set(float64(*d.Info.DocsRead))
		docsWrittenTotal.WithLabelValues(*d.DocID).Set(float64(*d.Info.DocsWritten))
		missingRevsFoundTotal.WithLabelValues(*d.DocID).Set(float64(*d.Info.MissingRevisionsFound))
		revsCheckedTotal.WithLabelValues(*d.DocID).Set(float64(*d.Info.RevisionsChecked))
	}
	return nil
}
