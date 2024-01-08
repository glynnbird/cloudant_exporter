package monitors

// ReplicationStatusCollector moves the data collection into a separate loop
// as it can be very time consuming (we use a 5s gap between pages in the
// retrieved results). We also only want to do this large-ish doc read every
// few minutes, so cache results for Prometheus's collector.

import (
	"log"
	"time"

	"github.com/IBM/cloudant-go-sdk/cloudantv1"
	"github.com/prometheus/client_golang/prometheus"
)

type ReplicationStatusMonitor struct {
	Cldt *cloudantv1.CloudantV1
}

var (
	replicatonStatus = prometheus.NewDesc(
		"cloudant_replication_status_count",
		"Current replication count by status",
		[]string{"status"}, nil,
	)
)

type ReplicationStatusCollector struct {
	Cldt         *cloudantv1.CloudantV1
	statusCounts map[string]uint
}

func NewReplicationStatusCollector(cldt *cloudantv1.CloudantV1) *ReplicationStatusCollector {
	return &ReplicationStatusCollector{
		Cldt: cldt,
		statusCounts: map[string]uint{
			"initializing": 0,
			"error":        0,
			"pending":      0,
			"running":      0,
			"crashing":     0,
			"completed":    0,
			"failed":       0,
		},
	}
}

// Start starts a ticker loop that runs forever to gather the metrics
func (cc ReplicationStatusCollector) Start() {
	var skip int = 0
	var batchSize int = 100
	var iterations = 0
	getSchedulerDocsOptions := cc.Cldt.NewGetSchedulerDocsOptions()
	getSchedulerDocsOptions.SetLimit(int64(batchSize))

	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		// repeat until we get a smaller batch than we asked for
		for {
			// fetch scheduler jobs
			getSchedulerDocsOptions.SetSkip(int64(skip))

			schedulerJobsResult, _, err := cc.Cldt.GetSchedulerDocs(getSchedulerDocsOptions)
			cloudantExporterHttpRequestTotal.WithLabelValues("replication_status").Inc()
			if err != nil {
				log.Printf("[ReplicationStatusCollector] Error retrieving replication status: %v", err)
				cloudantExporterHttpRequestErrorTotal.WithLabelValues("replication_status").Inc()
				break
			}
			for _, d := range schedulerJobsResult.Docs {
				cc.statusCounts[*d.State]++
			}
			skip += len(schedulerJobsResult.Docs)
			iterations++
			if len(schedulerJobsResult.Docs) < batchSize || iterations == 10 {
				break
			} else {
				time.Sleep(5 * time.Second)
			}
		}
	}
}

func (cc ReplicationStatusCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(cc, ch)
}

func (cc ReplicationStatusCollector) Collect(ch chan<- prometheus.Metric) {
	// output one metric per replication status
	for key, val := range cc.statusCounts {
		log.Printf("[ReplicationStatusCollector] %s %d", key, val)
		ch <- prometheus.MustNewConstMetric(
			replicatonStatus,
			prometheus.GaugeValue,
			float64(val),
			key,
		)
	}
}
