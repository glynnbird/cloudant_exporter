package monitors

import (
	"context"
	"encoding/json"
	"log"

	"github.com/IBM/cloudant-go-sdk/cloudantv1"
	"github.com/IBM/cloudant-go-sdk/common"
	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	throughputDesc = prometheus.NewDesc(
		"cloudant_throughput_current_req_per_second",
		"Current requests per second per class",
		[]string{"class", "ratelimited"}, nil,
	)
)

type ThroughputCollector struct {
	Cldt *cloudantv1.CloudantV1
}

func (cc ThroughputCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(cc, ch)
}

func (cc ThroughputCollector) Collect(ch chan<- prometheus.Metric) {

	tr, err := cc.ccmDiagnostics()
	if err != nil {
		log.Printf("[ThroughputCollector] Error retrieving CCM diagnostics: %v", err)
	}

	latest := tr.OperationHistory[len(tr.OperationHistory)-1]
	ch <- prometheus.MustNewConstMetric(
		throughputDesc,
		prometheus.GaugeValue,
		float64(latest.Lookup),
		"lookup", "false",
	)
	ch <- prometheus.MustNewConstMetric(
		throughputDesc,
		prometheus.GaugeValue,
		float64(latest.Write),
		"write", "false",
	)
	ch <- prometheus.MustNewConstMetric(
		throughputDesc,
		prometheus.GaugeValue,
		float64(latest.Query),
		"query", "false",
	)

	latest = tr.Deny429History[len(tr.Deny429History)-1]
	ch <- prometheus.MustNewConstMetric(
		throughputDesc,
		prometheus.GaugeValue,
		float64(latest.Lookup),
		"lookup", "true",
	)
	ch <- prometheus.MustNewConstMetric(
		throughputDesc,
		prometheus.GaugeValue,
		float64(latest.Write),
		"write", "true",
	)
	ch <- prometheus.MustNewConstMetric(
		throughputDesc,
		prometheus.GaugeValue,
		float64(latest.Query),
		"query", "true",
	)
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

func (tm *ThroughputCollector) ccmDiagnostics() (*ThroughputResponse, error) {
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
