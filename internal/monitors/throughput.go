package monitors

import (
	"encoding/json"
	"log"
	"time"

	"cloudant.com/couchmonitor/internal/utils"
	"github.com/IBM/cloudant-go-sdk/cloudantv1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type ThroughputMonitor struct {
	Cldt     *cloudantv1.CloudantV1
	Interval time.Duration
	FailBox  *utils.FailBox
}

var (
	throughput = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cloudant_throughput_current_req_per_second",
			Help: "Current requests per second per class",
		},
		[]string{"class"},
	)
)

func (tm *ThroughputMonitor) Name() string {
	return "ThroughputMonitor"
}

func (tm *ThroughputMonitor) Retrieve() error {
	getCurrentThroughputInformationOptions := tm.Cldt.NewGetCurrentThroughputInformationOptions()

	ti, _, err := tm.Cldt.GetCurrentThroughputInformation(getCurrentThroughputInformationOptions)
	if err != nil {
		return err
	}

	throughput.WithLabelValues("lookup").Set(float64(*ti.Throughput.Read))
	throughput.WithLabelValues("write").Set(float64(*ti.Throughput.Write))
	throughput.WithLabelValues("query").Set(float64(*ti.Throughput.Query))

	b, _ := json.Marshal(ti)
	log.Printf("[ThroughputMonitor] %v", string(b))

	return nil
}
