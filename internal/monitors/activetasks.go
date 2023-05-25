package monitors

import (
	"log"

	"cloudant.com/cloudant_exporter/internal/utils"
	"github.com/IBM/cloudant-go-sdk/cloudantv1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type ActiveTasksMonitor struct {
	Cldt *cloudantv1.CloudantV1
}

var (
	indexerChangesTotalGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cloudant_indexing_changes_total_documents",
		Help: "The total number of changes to index",
	},
		[]string{"node", "pid", "database", "design_document"},
	)
	indexerChangesDoneCounter = utils.AutoNewSettableCounterVec(prometheus.Opts{
		Name: "cloudant_indexing_changes_done_total",
		Help: "The total number of revisions processed by this indexer",
	},
		[]string{"node", "pid", "database", "design_document"},
	)
	compactionChangesTotalGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cloudant_compaction_changes_total_documents",
		Help: "The number of documents to compact",
	},
		[]string{"node", "pid", "database"},
	)
	compactionChangesDoneCounter = utils.AutoNewSettableCounterVec(prometheus.Opts{
		Name: "cloudant_compaction_changes_done_total",
		Help: "The total number of documents compacted by this compaction",
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
		switch *d.Type {
		case "indexer":
			log.Printf("[ActiveTasksMonitor] indexing ddoc %q db %q: changes %d", *d.DesignDocument, *d.Database, *d.TotalChanges)
			indexerChangesTotalGauge.WithLabelValues(*d.Node, *d.Pid, *d.Database, *d.DesignDocument).Set(float64(*d.TotalChanges))
			indexerChangesDoneCounter.WithLabelValues(*d.Node, *d.Pid, *d.Database, *d.DesignDocument).Set(float64(*d.ChangesDone))
		case "database_compaction":
			log.Printf("[ActiveTasksMonitor] compaction db %q total change %d done %d", *d.Database, *d.TotalChanges, *d.ChangesDone)
			compactionChangesTotalGauge.WithLabelValues(*d.Node, *d.Pid, *d.Database).Set(float64(*d.TotalChanges))
			compactionChangesDoneCounter.WithLabelValues(*d.Node, *d.Pid, *d.Database).Set(float64(*d.ChangesDone))
		default:
			// no prometheus output for replication, as that's handled by the ReplicationMonitor
		}
	}

	return nil
}
