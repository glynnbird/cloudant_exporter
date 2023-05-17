package monitors

import (
	"context"
	"encoding/json"
	"time"

	"cloudant.com/couchmonitor/internal/utils"
	"github.com/IBM/cloudant-go-sdk/cloudantv1"
	"github.com/IBM/cloudant-go-sdk/common"
	"github.com/IBM/go-sdk-core/v5/core"
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
		[]string{"class", "ratelimited"},
	)
)

func (tm *ThroughputMonitor) Name() string {
	return "ThroughputMonitor"
}

func (tm *ThroughputMonitor) Retrieve() error {
	tr, err := tm.ccmDiagnostics()
	if err != nil {
		return err
	}

	latest := tr.OperationHistory[len(tr.OperationHistory)-1]
	throughput.WithLabelValues("lookup", "false").Set(float64(latest.Lookup))
	throughput.WithLabelValues("write", "false").Set(float64(latest.Write))
	throughput.WithLabelValues("query", "false").Set(float64(latest.Query))

	latest = tr.Deny429History[len(tr.Deny429History)-1]
	throughput.WithLabelValues("lookup", "true").Set(float64(latest.Lookup))
	throughput.WithLabelValues("write", "true").Set(float64(latest.Write))
	throughput.WithLabelValues("query", "true").Set(float64(latest.Query))

	return nil
}

type ThroughputRecord struct {
	Ts     int64
	Lookup int64
	Write  int64
	Query  int64
}

type ThroughputResponse struct {
	Account          string
	Ts               int64
	Lookup           int64
	Write            int64
	Query            int64
	Deny429History   []ThroughputRecord
	OperationHistory []ThroughputRecord
}

func (tm *ThroughputMonitor) ccmDiagnostics() (*ThroughputResponse, error) {
	builder := core.NewRequestBuilder(core.GET)
	builder = builder.WithContext(context.Background())
	builder.EnableGzipCompression = tm.Cldt.GetEnableGzipCompression()
	_, err := builder.ResolveRequestURL(tm.Cldt.Service.Options.URL, `/_api/v2/user/ccm_diagnostics`, nil)
	if err != nil {
		return nil, err
	}

	sdkHeaders := common.GetSdkHeaders("cloudant", "V1", "GetCurrentThroughputInformation")
	for headerName, headerValue := range sdkHeaders {
		builder.AddHeader(headerName, headerValue)
	}
	builder.AddHeader("Accept", "application/json")

	request, err := builder.Build()
	if err != nil {
		return nil, err
	}

	var rawResponse json.RawMessage
	_, err = tm.Cldt.Service.Request(request, &rawResponse)
	if err != nil {
		return nil, err
	}
	tr := &ThroughputResponse{}
	if rawResponse != nil {
		err = json.Unmarshal(rawResponse, tr)
		if err != nil {
			return nil, err
		}
	}

	return tr, nil
}
