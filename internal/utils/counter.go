package utils

import (
	"sort"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"google.golang.org/protobuf/proto"
)

// LabelPairSorter implements sort.Interface. It is used to sort a slice of
// dto.LabelPair pointers.
// Copied from "github.com/prometheus/client_golang/prometheus"
type LabelPairSorter []*dto.LabelPair

func (s LabelPairSorter) Len() int {
	return len(s)
}

func (s LabelPairSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s LabelPairSorter) Less(i, j int) bool {
	return s[i].GetName() < s[j].GetName()
}

// SettableCounter implements a Prometheus counter-type metric
// that can be directly set by a user, in contrast to the builtin
// prometheus.Counter which can only be incremented.
//
// This allows us to more effectively "proxy" the counter type
// metrics that we receive from CouchDB, otherwise we have to
// use a prometheus.Gauge, which doesn't give good affordances
// within Prometheus itself, eg, calculating rates.
type SettableCounter struct {
	val float64

	desc       *prometheus.Desc
	labelPairs []*dto.LabelPair
}

// NewSettableCounter returns a new SettableCounter.
func NewSettableCounter(opts prometheus.Opts) *SettableCounter {
	desc := prometheus.NewDesc(
		prometheus.BuildFQName(opts.Namespace, opts.Subsystem, opts.Name),
		opts.Help,
		nil,
		opts.ConstLabels,
	)

	// Taken from NewDesc implementation
	constLabelPairs := make([]*dto.LabelPair, 0, len(opts.ConstLabels))
	for n, v := range opts.ConstLabels {
		constLabelPairs = append(constLabelPairs, &dto.LabelPair{
			Name:  proto.String(n),
			Value: proto.String(v),
		})
	}
	sort.Sort(LabelPairSorter(constLabelPairs))

	return &SettableCounter{
		desc:       desc,
		labelPairs: constLabelPairs,
	}
}

// Desc implements prometheus.Metric
func (c *SettableCounter) Desc() *prometheus.Desc {
	return c.desc
}

// Write implements prometheus.Metric
func (c *SettableCounter) Write(out *dto.Metric) error {
	out.Label = c.labelPairs
	out.Counter = &dto.Counter{Value: proto.Float64(c.val), Exemplar: nil}
	return nil
}

// Set directly sets the counter to v.
func (c *SettableCounter) Set(v float64) {
	c.val = v
}

// SettableCounterVec creates a metric vector containing SettableCounter.
//
// We do this by customising prometheus.MetricVec. We have to create a few methods to
// support this, per the prometheus.MetricVec documentation:
//
// > To create a FooVec for custom Metric Foo, embed a pointer to MetricVec in
// > FooVec and initialize it with NewMetricVec. Implement wrappers for
// > GetMetricWithLabelValues and GetMetricWith that return (Foo, error) rather
// > than (Metric, error). Similarly, create a wrapper for CurryWith that returns
// > (*FooVec, error) rather than (*MetricVec, error). It is recommended to also
// > add the convenience methods WithLabelValues, With, and MustCurryWith, which
// > panic instead of returning errors. See also the MetricVec example.
type SettableCounterVec struct {
	*prometheus.MetricVec
}

// NewSettableCounterVec creates a new SettableCounterVec.
func NewSettableCounterVec(opts prometheus.Opts, labelNames []string) *SettableCounterVec {
	desc := prometheus.V2.NewDesc(
		prometheus.BuildFQName(opts.Namespace, opts.Subsystem, opts.Name),
		opts.Help,
		prometheus.UnconstrainedLabels(labelNames),
		opts.ConstLabels,
	)
	return &SettableCounterVec{
		MetricVec: prometheus.NewMetricVec(desc, func(lvs ...string) prometheus.Metric {
			return &SettableCounter{
				desc:       desc,
				labelPairs: prometheus.MakeLabelPairs(desc, lvs),
			}
		}),
	}
}

// AutoNewSettableCounterVec creates and MustRegisters() a new SettableCounterVec.
func AutoNewSettableCounterVec(opts prometheus.Opts, labelNames []string) *SettableCounterVec {
	v := NewSettableCounterVec(opts, labelNames)
	prometheus.DefaultRegisterer.MustRegister(v)
	return v
}

// GetMetricWithLabelValues is required to customise MetricVec (see SettableCounterVec doc)
func (v *SettableCounterVec) GetMetricWithLabelValues(lvs ...string) (*SettableCounter, error) {
	metric, err := v.MetricVec.GetMetricWithLabelValues(lvs...)
	if metric != nil {
		return metric.(*SettableCounter), err
	}
	return nil, err
}

// GetMetricWith is required to customise MetricVec (see SettableCounterVec doc)
func (v *SettableCounterVec) GetMetricWith(labels prometheus.Labels) (*SettableCounter, error) {
	metric, err := v.MetricVec.GetMetricWith(labels)
	if metric != nil {
		return metric.(*SettableCounter), err
	}
	return nil, err
}

// CurryWith is required to customise MetricVec (see SettableCounterVec doc)
func (v *SettableCounterVec) CurryWith(labels prometheus.Labels) (*SettableCounterVec, error) {
	vec, err := v.MetricVec.CurryWith(labels)
	if vec != nil {
		return &SettableCounterVec{vec}, err
	}
	return nil, err
}

// MustCurryWith works as CurryWith but panics where CurryWith would have
// returned an error.
func (v *SettableCounterVec) MustCurryWith(labels prometheus.Labels) *SettableCounterVec {
	vec, err := v.CurryWith(labels)
	if err != nil {
		panic(err)
	}
	return vec
}

// WithLabelValues works as GetMetricWithLabelValues, but panics where
// GetMetricWithLabelValues would have returned an error. Not returning an
// error allows shortcuts like
//
//	myVec.WithLabelValues("404", "GET").Add(42)
func (v *SettableCounterVec) WithLabelValues(lvs ...string) *SettableCounter {
	c, err := v.GetMetricWithLabelValues(lvs...)
	if err != nil {
		panic(err)
	}
	return c
}
