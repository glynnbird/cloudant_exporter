package monitors

import (
	"log"
	"time"

	"cloudant.com/cloudant_exporter/internal/utils"
	"github.com/IBM/cloudant-go-sdk/cloudantv1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type ReplicationStatusMonitor struct {
	Cldt     *cloudantv1.CloudantV1
	Interval time.Duration
	FailBox  *utils.FailBox
}

var (
	replicatonStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cloudant_replication_status_count",
			Help: "Current replication count by status",
		},
		[]string{"status"},
	)
)

func (rc *ReplicationStatusMonitor) Name() string {
	return "ReplicationStatusMonitor"
}

func (rc *ReplicationStatusMonitor) Retrieve() error {
	var skip int = 0
	var batchSize int = 100
	var iterations = 0
	getSchedulerDocsOptions := rc.Cldt.NewGetSchedulerDocsOptions()
	getSchedulerDocsOptions.SetLimit(int64(batchSize))
	statusCounts := map[string]uint{
		"initializing": 0,
		"error":        0,
		"pending":      0,
		"running":      0,
		"crashing":     0,
		"completed":    0,
		"failed":       0,
	}

	// repeat until we get a smaller batch than we asked for
	for {
		// fetch scheduler jobs
		getSchedulerDocsOptions.SetSkip(int64(skip))

		schedulerJobsResult, _, err := rc.Cldt.GetSchedulerDocs(getSchedulerDocsOptions)
		if err != nil {
			return err
		}
		for _, d := range schedulerJobsResult.Docs {
			statusCounts[*d.State]++
		}
		skip += len(schedulerJobsResult.Docs)
		iterations++
		if len(schedulerJobsResult.Docs) < batchSize || iterations == 10 {
			break
		} else {
			time.Sleep(5 * time.Second)
		}
	}

	// output one metric per replication status
	for key, val := range statusCounts {
		log.Printf("[ReplicationProgressMonitor] %s %d", key, val)
		replicatonStatus.WithLabelValues(key).Set(float64(val))
	}

	return nil
}
