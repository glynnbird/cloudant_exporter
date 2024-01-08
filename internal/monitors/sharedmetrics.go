package monitors

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	cloudantExporterHttpRequestTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cloudant_exporter_http_request_total",
		Help: "Total HTTP requests made to Cloudant service",
	},
		[]string{"collector"},
	)
	cloudantExporterHttpRequestErrorTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cloudant_exporter_http_request_error_total",
		Help: "Total errors for HTTP requests to Cloudant service",
	},
		[]string{"collector"},
	)
)
