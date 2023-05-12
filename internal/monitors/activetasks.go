package monitors

import (
	"log"
	"time"

	"github.com/IBM/cloudant-go-sdk/cloudantv1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type ActiveTasksMonitor struct {
	Cldt     *cloudantv1.CloudantV1
	Interval time.Duration
	Done     chan bool
}

var (
	indexerChangesTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cloudant_indexing_changes_total",
		Help: "The total number of changes to index",
	},
		[]string{"database", "design_document"},
	)
	indexerChangesDone = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cloudant_indexing_changes_done",
		Help: "The  number of changes indexed",
	},
		[]string{"database", "design_document"},
	)
	compactionChangesTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cloudant_compaction_changes_total",
		Help: "The number of documents to compact",
	},
		[]string{"database"},
	)
	compactionChangesDone = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cloudant_compaction_changes_done",
		Help: "The number of documents to compacted",
	},
		[]string{"database"},
	)
)

func (rc *ActiveTasksMonitor) Go() {
	ticker := time.NewTicker(rc.Interval)

	go func() {
		for {
			select {
			case <-rc.Done:
				return
			case t := <-ticker.C:
				log.Println("Polling Cloudant active tasks", t)

				// fetch active tasks
				getActiveTasksOptions := rc.Cldt.NewGetActiveTasksOptions()
				activeTaskResult, _, err := rc.Cldt.GetActiveTasks(getActiveTasksOptions)

				if err != nil {
					log.Printf("ActiveTasksMonitor: Error in GetActiveTasks: %v", err)
					continue
				}
				for _, d := range activeTaskResult {
					if *d.Type == "indexer" {
						log.Printf("Active Tasks: indexing ddoc %q db %q: changes %d", *d.DesignDocument, *d.Database, *d.TotalChanges)
						indexerChangesTotal.WithLabelValues(*d.Database, *d.DesignDocument).Set(float64(*d.TotalChanges))
						indexerChangesDone.WithLabelValues(*d.Database, *d.DesignDocument).Set(float64(*d.ChangesDone))
					}
					if *d.Type == "replication" {
						log.Printf("Active Tasks: replication %q", *d.DocID)
						// no prometheus output for replication, as that's handled elsewhere
					}
					if *d.Type == "database_compaction" {
						log.Printf("Active Tasks: compaction db %q total change %d done %d", *d.Database, *d.TotalChanges, *d.ChangesDone)
						compactionChangesTotal.WithLabelValues(*d.Database).Set(float64(*d.TotalChanges))
						compactionChangesTotal.WithLabelValues(*d.Database).Set(float64(*d.ChangesDone))
					}
				}
			}
		}
	}()

}
