package main

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func setUpGaugeTest(t *testing.T) *GaugeRegistry {
	mockOut := `              capacity     operations     bandwidth
	pool        alloc   free   read  write   read  write
	----------  -----  -----  -----  -----  -----  -----
	tank         200M   792M      0      0      0    310
	test0       94.5K  79.9M      0      0    152    539
	----------  -----  -----  -----  -----  -----  -----
	`
	hostname, err := Hostname()
	if err != nil {
		t.Fatal(err)
	}
	zpools, err := ParseZPoolIOStat(mockOut, hostname)
	if err != nil {
		t.Fatal(err)
	}
	gr := NewGaugeRegistry("")
	gr.PrometheusRegistry = prometheus.NewRegistry()
	for _, zpool := range zpools {
		gr.Register(zpool)
	}

	return gr
}

func TestUpdate(t *testing.T) {
	// TODO find a better way to test this
	mockOutNew := `              capacity     operations     bandwidth
	pool        alloc   free   read  write   read  write
	----------  -----  -----  -----  -----  -----  -----
	tank         300M   923M      1      1      1    310
	test0       10.5K  65.9M      1      1    152    539
	----------  -----  -----  -----  -----  -----  -----
	`
	host, err := Hostname()
	if err != nil {
		t.Fatal(err)
	}
	gr := setUpGaugeTest(t)
	m1 := make(map[string]*prometheus.Gauge)

	zpools, err := ParseZPoolIOStat(mockOutNew, host)
	for k, v := range gr.Gauges {
		m1[k] = v
	}
	for _, zpool := range zpools {
		gr.Update(zpool)
	}
}
