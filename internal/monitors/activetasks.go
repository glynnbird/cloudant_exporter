package monitors

import (
	"log"

	"github.com/IBM/cloudant-go-sdk/cloudantv1"
	"github.com/prometheus/client_golang/prometheus"
)

type ActiveTasksMonitor struct {
	Cldt *cloudantv1.CloudantV1
}

var (
	indexerChangesTotalGaugeDesc = prometheus.NewDesc(
		"cloudant_indexing_changes_total_documents",
		"The total number of changes to index",
		[]string{"node", "pid", "database", "design_document"}, nil,
	)
	indexerChangesDoneCounterDesc = prometheus.NewDesc(
		"cloudant_indexing_changes_done_total",
		"The total number of revisions processed by this indexer",
		[]string{"node", "pid", "database", "design_document"}, nil,
	)
	compactionChangesTotalGaugeDesc = prometheus.NewDesc(
		"cloudant_compaction_changes_total_documents",
		"The number of documents to compact",
		[]string{"node", "pid", "database"}, nil,
	)
	compactionChangesDoneCounterDesc = prometheus.NewDesc(
		"cloudant_compaction_changes_done_total",
		"The total number of documents compacted by this compaction",
		[]string{"node", "pid", "database"}, nil,
	)
)

type ActiveTasksCollector struct {
	Cldt *cloudantv1.CloudantV1
}

func (cc ActiveTasksCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(cc, ch)
}

func (rc ActiveTasksCollector) Collect(ch chan<- prometheus.Metric) {
	// fetch active tasks
	getActiveTasksOptions := rc.Cldt.NewGetActiveTasksOptions()
	activeTaskResult, _, err := rc.Cldt.GetActiveTasks(getActiveTasksOptions)
	cloudantExporterHttpRequestTotal.WithLabelValues("active_tasks").Inc()

	if err != nil {
		log.Printf("[ActiveTasksCollector] Error retrieving active tasks: %v", err)
		cloudantExporterHttpRequestErrorTotal.WithLabelValues("active_tasks").Inc()
		return
	}

	for _, d := range activeTaskResult {
		switch *d.Type {
		case "indexer":
			log.Printf("[ActiveTasksMonitor] indexing ddoc %q db %q: changes %d", *d.DesignDocument, *d.Database, *d.TotalChanges)
			ch <- prometheus.MustNewConstMetric(
				indexerChangesTotalGaugeDesc,
				prometheus.GaugeValue,
				float64(*d.TotalChanges),
				*d.Node, *d.Pid, *d.Database, *d.DesignDocument,
			)
			ch <- prometheus.MustNewConstMetric(
				indexerChangesDoneCounterDesc,
				prometheus.CounterValue,
				float64(*d.ChangesDone),
				*d.Node, *d.Pid, *d.Database, *d.DesignDocument,
			)
		case "database_compaction":
			log.Printf("[ActiveTasksMonitor] compaction db %q total change %d done %d", *d.Database, *d.TotalChanges, *d.ChangesDone)
			ch <- prometheus.MustNewConstMetric(
				compactionChangesTotalGaugeDesc,
				prometheus.GaugeValue,
				float64(*d.TotalChanges),
				*d.Node, *d.Pid, *d.Database, *d.DesignDocument,
			)
			ch <- prometheus.MustNewConstMetric(
				compactionChangesDoneCounterDesc,
				prometheus.CounterValue,
				float64(*d.ChangesDone),
				*d.Node, *d.Pid, *d.Database, *d.DesignDocument,
			)
		default:
			// no prometheus output for replication, as that's handled by the ReplicationMonitor
		}
	}
}
