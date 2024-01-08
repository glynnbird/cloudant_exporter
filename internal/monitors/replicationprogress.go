package monitors

import (
	"log"

	"github.com/IBM/cloudant-go-sdk/cloudantv1"
	"github.com/prometheus/client_golang/prometheus"
)

type ReplicationProgressMonitor struct {
	Cldt *cloudantv1.CloudantV1
}

var (
	// Changes pending mostly goes down, but can go up if the replication
	// begins to fall behind. It's definitely a gauge.
	changesPendingTotalDesc = prometheus.NewDesc(
		"cloudant_replication_changes_pending_total",
		"The number of changes remaining to process (approximately)",
		[]string{"docid"}, nil,
	)

	// Everything else is a counter-type, even if it's reset to zero somehow,
	// at least if we are correctly labelling the metric.
	docsReadTotalDesc = prometheus.NewDesc(
		"cloudant_replication_docs_read_total",
		"Total number of documents read from the source database",
		[]string{"docid"}, nil,
	)
	docsWrittenTotalDesc = prometheus.NewDesc(
		"cloudant_replication_docs_written_total",
		"Total number of documents written to the target database",
		[]string{"docid"}, nil,
	)
	docWriteFailuresTotalDesc = prometheus.NewDesc(
		"cloudant_replication_doc_write_failures_total",
		"The number of failures writing documents to the target",
		[]string{"docid"}, nil,
	)
	missingRevsFoundTotalDesc = prometheus.NewDesc(
		"cloudant_replication_missing_revs_found_total",
		"Total number of revs found so far on the source that are not at the target",
		[]string{"docid"}, nil,
	)
	revsCheckedTotalDesc = prometheus.NewDesc(
		"cloudant_replication_revs_checked_total",
		"Total number of revs processed on the source",
		[]string{"docid"}, nil,
	)
)

type ReplicationProgressCollector struct {
	Cldt *cloudantv1.CloudantV1
}

func (cc ReplicationProgressCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(cc, ch)
}

func (cc ReplicationProgressCollector) Collect(ch chan<- prometheus.Metric) {
	// fetch scheduler status
	getSchedulerDocsOptions := cc.Cldt.NewGetSchedulerDocsOptions()
	getSchedulerDocsOptions.SetLimit(50)
	getSchedulerDocsOptions.SetStates([]string{"running"})

	schedulerDocsResult, _, err := cc.Cldt.GetSchedulerDocs(getSchedulerDocsOptions)
	cloudantExporterHttpRequestTotal.WithLabelValues("replication_progress").Inc()
	if err != nil {
		log.Printf("[ReplicationProgressCollector] Error retrieving replication progress: %v", err)
		cloudantExporterHttpRequestErrorTotal.WithLabelValues("replication_progress").Inc()
		return
	}
	for _, d := range schedulerDocsResult.Docs {
		if d.DocID == nil || d.Info == nil {
			continue
		}
		log.Printf("[ReplicationProgressMonitor] Replication %q: docs written %d", *d.DocID, *d.Info.DocsWritten)

		if d.Info.ChangesPending != nil {
			ch <- prometheus.MustNewConstMetric(
				changesPendingTotalDesc,
				prometheus.GaugeValue,
				float64(*d.Info.ChangesPending),
				*d.DocID,
			)
		}
		ch <- prometheus.MustNewConstMetric(
			docWriteFailuresTotalDesc,
			prometheus.CounterValue,
			float64(*d.Info.DocWriteFailures),
			*d.DocID,
		)
		ch <- prometheus.MustNewConstMetric(
			docsReadTotalDesc,
			prometheus.CounterValue,
			float64(*d.Info.DocsRead),
			*d.DocID,
		)
		ch <- prometheus.MustNewConstMetric(
			docsWrittenTotalDesc,
			prometheus.CounterValue,
			float64(*d.Info.DocsWritten),
			*d.DocID,
		)
		ch <- prometheus.MustNewConstMetric(
			missingRevsFoundTotalDesc,
			prometheus.CounterValue,
			float64(*d.Info.MissingRevisionsFound),
			*d.DocID,
		)
		ch <- prometheus.MustNewConstMetric(
			revsCheckedTotalDesc,
			prometheus.CounterValue,
			float64(*d.Info.RevisionsChecked),
			*d.DocID,
		)
	}
}
