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
	// dummy labels (for now)
	labels := make(map[string]string)
	labels["source"] = "https://4268d9ec-250f-4d59-bcbd-fb47a14ef856-bluemix.cloudant.com/americanairlines/"
	labels["target"] = "https://4268d9ec-250f-4d59-bcbd-fb47a14ef856-bluemix.cloudant.com/aa3/"
	labels["id"] = "728b1e5bf539e8a1ee796d25869c5b37"

	docsProcessed := prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "replication_docs_processed_total",
		Help:        "The number of documents written to the target",
		ConstLabels: prometheus.Labels(labels),
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
				log.Println("Polling Cloudant replication", t)

				// fetch scheduler status
				getSchedulerDocsOptions := rc.Cldt.NewGetSchedulerDocsOptions()
				schedulerDocsResult, _, _ := rc.Cldt.GetSchedulerDocs(getSchedulerDocsOptions)

				// to stdout - not plumbed into Prometheus client yet
				if len(schedulerDocsResult.Docs) > 0 {
					log.Printf("docs written %d", *schedulerDocsResult.Docs[0].Info.DocsWritten)
					docsProcessed.Set(float64(*schedulerDocsResult.Docs[0].Info.DocsWritten))

				}
			}
		}
	}()

}
