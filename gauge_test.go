package main

import (
	"fmt"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestRegister(t *testing.T) {
	mockOut := `              capacity     operations     bandwidth
	pool        alloc   free   read  write   read  write
	----------  -----  -----  -----  -----  -----  -----
	tank         200M   792M      0      0      0    310
	test0       94.5K  79.9M      0      0    152    539
	----------  -----  -----  -----  -----  -----  -----
	`
	hostname := hostname(t)
	zpools, err := ParseZPoolIOStat(mockOut, hostname)
	if err != nil {
		t.Fatal(err)
	}
	gr := NewGaugeRegistry("")
	gr.PrometheusRegistry = prometheus.NewRegistry()

	gr.Register(zpools[0])

	if len(gr.gauges) == 0 {
		t.Error("No gauges were registered")
	}
}

func TestParseZPoolIOStat(t *testing.T) {
	// only 1 zpool
	mockOut := `              capacity     operations     bandwidth
	pool        alloc   free   read  write   read  write
	----------  -----  -----  -----  -----  -----  -----
	tank         200M   792M      0      0      0    310
	----------  -----  -----  -----  -----  -----  -----
	`
	zpools, err := ParseZPoolIOStat(mockOut, hostname(t))
	if err != nil {
		t.Fatal(err)
	}

	if len(zpools) != 1 {
		t.Fatalf("Incorrect number of zpool iostat parsed. Want 2, got %d", len(zpools))
	}

	for _, zpool := range zpools {
		zpoolsType := fmt.Sprintf("%T", zpool)
		if zpoolsType != "main.ZPool" {
			t.Errorf("Should have been type main.ZPool, not %s", zpoolsType)
		}
	}

	// 2 zpools
	mockOut = `              capacity     operations     bandwidth
	pool        alloc   free   read  write   read  write
	----------  -----  -----  -----  -----  -----  -----
	tank         200M   792M      0      0      0    310
	test0       94.5K  79.9M      0      0    152    539
	----------  -----  -----  -----  -----  -----  -----
	`
	zpools, err = ParseZPoolIOStat(mockOut, hostname(t))
	if err != nil {
		t.Fatal(err)
	}

	if len(zpools) != 2 {
		t.Fatalf("Incorrect number of zpool iostat parsed. Want 2, got %d", len(zpools))
	}

	for _, zpool := range zpools {
		zpoolsType := fmt.Sprintf("%T", zpool)
		if zpoolsType != "main.ZPool" {
			t.Errorf("Should have been type main.ZPool, not %s", zpoolsType)
		}
	}
}
