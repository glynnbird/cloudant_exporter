package monitors

import (
	"log"
	"time"

	"github.com/IBM/cloudant-go-sdk/cloudantv1"
	"github.com/prometheus/client_golang/prometheus"
)

type ReplicationCollector struct {
	Reg      *prometheus.Registry
	Cldt     *cloudantv1.CloudantV1
	Interval time.Duration
	Done     chan bool
}

func (rc *ReplicationCollector) Go() {
	docsProcessed := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "replication_docs_processed_total",
	})
	rc.Reg.MustRegister(docsProcessed)

	ticker := time.NewTicker(rc.Interval)

	go func() {
		for {
			select {
			case <-rc.Done:
				return
			case t := <-ticker.C:
				log.Println("Tick at", t)
				docsProcessed.Inc()
			}
		}
	}()

}
