package prom

import (
	"math"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// Prom struct info
type Prom struct {
	Counter *prometheus.GaugeVec
	State   *prometheus.GaugeVec
}

// New creates a Prom instance.
func New() *Prom {
	return &Prom{}
}

// WithCounter sets counter.
func (p *Prom) WithCounter(name string, labels []string) *Prom {
	if p == nil || p.Counter != nil {
		return p
	}
	p.Counter = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: name,
			Help: name,
		}, labels)
	prometheus.MustRegister(p.Counter)
	return p
}

// WithState sets state.
func (p *Prom) WithState(name string, labels []string) *Prom {
	if p == nil || p.State != nil {
		return p
	}
	p.State = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: name,
			Help: name,
		}, labels)
	prometheus.MustRegister(p.State)
	return p
}

// Incr increments one stat counter without sampling
func (p *Prom) Incr(name string, extra ...string) {
	label := append([]string{name}, extra...)
	if p.Counter != nil {
		p.Counter.WithLabelValues(label...).Inc()
	}
}

// Decr decrements one stat counter without sampling
func (p *Prom) Decr(name string, extra ...string) {
	if p.Counter != nil {
		label := append([]string{name}, extra...)
		p.Counter.WithLabelValues(label...).Desc()
	}
}

// Add add count    v must > 0
func (p *Prom) Add(name string, v int64, extra ...string) {
	label := append([]string{name}, extra...)
	if p.Counter != nil {
		p.Counter.WithLabelValues(label...).Add(float64(v))
	}
}

func ToFloat64(m *dto.Metric) float64 {
	if m.Gauge != nil {
		return m.Gauge.GetValue()
	}
	if m.Counter != nil {
		return m.Counter.GetValue()
	}
	if m.Untyped != nil {
		return m.Untyped.GetValue()
	}
	return math.NaN()
}

func FindMetricFamily(name string) (*dto.MetricFamily, bool) {
	gather := prometheus.DefaultRegisterer.(prometheus.Gatherer)
	mfs, _ := gather.Gather()
	for _, mf := range mfs {
		if mf.GetName() == name {
			return mf, true
		}
	}
	return nil, false
}

func GetLabel(m *dto.Metric, name string) (*dto.LabelPair, bool) {
	for _, lv := range m.GetLabel() {
		if lv.GetName() == name {
			return lv, true
		}
	}
	return nil, false
}
