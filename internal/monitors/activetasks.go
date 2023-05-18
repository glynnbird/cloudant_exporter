package monitors

import (
	"log"
	"time"

	"cloudant.com/couchmonitor/internal/utils"
	"github.com/IBM/cloudant-go-sdk/cloudantv1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type ActiveTasksMonitor struct {
	Cldt     *cloudantv1.CloudantV1
	Interval time.Duration
	FailBox  *utils.FailBox
}

var (
	indexerChangesTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cloudant_indexing_changes_total",
		Help: "The total number of changes to index",
	},
		[]string{"node", "pid", "database", "design_document"},
	)
	indexerChangesDone = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cloudant_indexing_changes_done",
		Help: "The  number of changes indexed",
	},
		[]string{"node", "pid", "database", "design_document"},
	)
	compactionChangesTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cloudant_compaction_changes_total",
		Help: "The number of documents to compact",
	},
		[]string{"node", "pid", "database"},
	)
	compactionChangesDone = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cloudant_compaction_changes_done",
		Help: "The number of documents to compacted",
	},
		[]string{"node", "pid", "database"},
	)
)

func (rc *ActiveTasksMonitor) Name() string {
	return "ActiveTasksMonitor"
}

func (rc *ActiveTasksMonitor) Retrieve() error {
	// fetch active tasks
	getActiveTasksOptions := rc.Cldt.NewGetActiveTasksOptions()
	activeTaskResult, _, err := rc.Cldt.GetActiveTasks(getActiveTasksOptions)

	if err != nil {
		return err
	}

	for _, d := range activeTaskResult {
		if *d.Type == "indexer" {
			log.Printf("[ActiveTasksMonitor] indexing ddoc %q db %q: changes %d", *d.DesignDocument, *d.Database, *d.TotalChanges)
			indexerChangesTotal.WithLabelValues(*d.Node, *d.Pid, *d.Database, *d.DesignDocument).Set(float64(*d.TotalChanges))
			indexerChangesDone.WithLabelValues(*d.Node, *d.Pid, *d.Database, *d.DesignDocument).Set(float64(*d.ChangesDone))
		}
		if *d.Type == "replication" {
			// no prometheus output for replication, as that's handled by the ReplicationMonitor
		}
		if *d.Type == "database_compaction" {
			log.Printf("[ActiveTasksMonitor] compaction db %q total change %d done %d", *d.Database, *d.TotalChanges, *d.ChangesDone)
			compactionChangesTotal.WithLabelValues(*d.Node, *d.Pid, *d.Database).Set(float64(*d.TotalChanges))
			compactionChangesDone.WithLabelValues(*d.Node, *d.Pid, *d.Database).Set(float64(*d.ChangesDone))
		}
	}

	return nil
}
