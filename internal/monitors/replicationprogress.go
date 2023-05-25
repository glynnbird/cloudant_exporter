package monitors

import (
	"log"

	"cloudant.com/cloudant_exporter/internal/utils"
	"github.com/IBM/cloudant-go-sdk/cloudantv1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type ReplicationProgressMonitor struct {
	Cldt *cloudantv1.CloudantV1
}

var (
	// Changes pending mostly goes down, but can go up if the replication
	// begins to fall behind. It's definitely a gauge.
	changesPendingTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cloudant_replication_changes_pending_total",
		Help: "The number of changes remaining to process (approximately)",
	},
		[]string{"docid"},
	)

	// Everything else is a counter-type, even if it's reset to zero somehow,
	// at least if we are correctly labelling the metric.
	docWriteFailuresTotal = utils.AutoNewSettableCounterVec(prometheus.Opts{
		Name: "cloudant_replication_doc_write_failures_total",
		Help: "The number of failures writing documents to the target",
	},
		[]string{"docid"},
	)
	docsReadTotal = utils.AutoNewSettableCounterVec(prometheus.Opts{
		Name: "cloudant_replication_docs_read_total",
		Help: "Total number of documents read from the source database",
	},
		[]string{"docid"},
	)
	docsWrittenTotal = utils.AutoNewSettableCounterVec(prometheus.Opts{
		Name: "cloudant_replication_docs_written_total",
		Help: "Total number of documents written to the target database",
	},
		[]string{"docid"},
	)
	missingRevsFoundTotal = utils.AutoNewSettableCounterVec(prometheus.Opts{
		Name: "cloudant_replication_missing_revs_found_total",
		Help: "Total number of revs found so far on the source that are not at the target",
	},
		[]string{"docid"},
	)
	revsCheckedTotal = utils.AutoNewSettableCounterVec(prometheus.Opts{
		Name: "cloudant_replication_revs_checked_total",
		Help: "Total number of revs processed on the source",
	},
		[]string{"docid"},
	)
)

func (rc *ReplicationProgressMonitor) Name() string {
	return "ReplicationProgressMonitor"
}

func (rc *ReplicationProgressMonitor) Retrieve() error {
	// fetch scheduler status
	getSchedulerDocsOptions := rc.Cldt.NewGetSchedulerDocsOptions()
	getSchedulerDocsOptions.SetLimit(50)
	getSchedulerDocsOptions.SetStates([]string{"running"})

	schedulerDocsResult, _, err := rc.Cldt.GetSchedulerDocs(getSchedulerDocsOptions)
	if err != nil {
		return err
	}
	for _, d := range schedulerDocsResult.Docs {
		log.Printf("[ReplicationProgressMonitor] Replication %q: docs written %d", *d.DocID, *d.Info.DocsWritten)
		if d.Info.ChangesPending != nil {
			changesPendingTotal.WithLabelValues(*d.DocID).Set(float64(*d.Info.ChangesPending))
		}
		docWriteFailuresTotal.WithLabelValues(*d.DocID).Set(float64(*d.Info.DocWriteFailures))
		docsReadTotal.WithLabelValues(*d.DocID).Set(float64(*d.Info.DocsRead))
		docsWrittenTotal.WithLabelValues(*d.DocID).Set(float64(*d.Info.DocsWritten))
		missingRevsFoundTotal.WithLabelValues(*d.DocID).Set(float64(*d.Info.MissingRevisionsFound))
		revsCheckedTotal.WithLabelValues(*d.DocID).Set(float64(*d.Info.RevisionsChecked))
	}
	return nil
}
