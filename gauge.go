package main

import (
	"sort"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func KeyFromMap(name string, m map[string]string) string {
	var s string
	s = s + name
	// maps are unordered, so sort by key
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// create a unique key based off name + labels
	for _, k := range keys {
		s = s + k + m[k]
	}
	return s
}

// Facade around registering & updating exports
type GaugeRegistry struct {
	Gauges             map[string]*prometheus.Gauge
	Hostname           string
	PrometheusRegistry *prometheus.Registry
	Namespace          string
}

func NewGaugeRegistry(prometheusNamespace string) *GaugeRegistry {
	gr := GaugeRegistry{}
	gr.Gauges = make(map[string]*prometheus.Gauge)
	gr.Namespace = prometheusNamespace
	return &gr
}

type Stat struct {
	Name   string
	Help   string
	Val    float64
	Labels map[string]string
}

func (gr *GaugeRegistry) Register(z ZPool) {
	// https://stackoverflow.com/questions/18926303/iterate-through-the-fields-of-a-struct-in-go
	gr.registerGauge(z.CapAlloc)
	gr.registerGauge(z.CapFree)
	gr.registerGauge(z.OpsRead)
	gr.registerGauge(z.OpsWrite)
	gr.registerGauge(z.BdwRead)
	gr.registerGauge(z.BdwWrite)
}

func (gr *GaugeRegistry) updateStat(s Stat) {
	key := KeyFromMap(s.Name, s.Labels)
	gauge := *gr.Gauges[key]
	gauge.Set(s.Val)
}

func (gr *GaugeRegistry) Update(z ZPool) {
	gr.updateStat(z.CapAlloc)
	gr.updateStat(z.CapFree)
	gr.updateStat(z.OpsRead)
	gr.updateStat(z.OpsWrite)
	gr.updateStat(z.BdwRead)
	gr.updateStat(z.BdwWrite)
}

func (gr *GaugeRegistry) registerGauge(s Stat) {
	g := promauto.NewGauge(prometheus.GaugeOpts{
		Name:        s.Name,
		Namespace:   gr.Namespace,
		Help:        s.Help,
		ConstLabels: s.Labels,
	})
	key := KeyFromMap(s.Name, s.Labels)
	gr.Gauges[key] = &g

	g.Set(s.Val)
	gr.PrometheusRegistry.MustRegister(g)
}
