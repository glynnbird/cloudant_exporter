package monitors

import (
	"encoding/json"
	"log"
	"time"

	"github.com/IBM/cloudant-go-sdk/cloudantv1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type ThroughputMonitor struct {
	Cldt     *cloudantv1.CloudantV1
	Interval time.Duration
	Done     chan bool
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

func (tm *ThroughputMonitor) Go() {
	ticker := time.NewTicker(tm.Interval)
	go func() {
		for {
			select {
			case <-tm.Done:
				return
			case <-ticker.C:
				log.Println("ThroughputMonitor ticker")
				tm.tick()
			}
		}
	}()
}

func (tm *ThroughputMonitor) tick() {
	getCurrentThroughputInformationOptions := tm.Cldt.NewGetCurrentThroughputInformationOptions()

	ti, _, err := tm.Cldt.GetCurrentThroughputInformation(getCurrentThroughputInformationOptions)
	if err != nil {
		log.Printf("ThroughputMonitor error getting throughput: %v", err)
		return
	}

	throughput.WithLabelValues("lookup").Set(float64(*ti.Throughput.Read))
	throughput.WithLabelValues("write").Set(float64(*ti.Throughput.Write))
	throughput.WithLabelValues("query").Set(float64(*ti.Throughput.Query))

	b, _ := json.Marshal(ti)
	log.Println(string(b))

}
