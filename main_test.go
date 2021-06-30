package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func hostname(t *testing.T) (hostname string) {
	hostname, err := os.Hostname()
	if err != nil {
		t.Fatal(err)
	}
	return
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
func TestSizeToBytes(t *testing.T) {
	cases := []struct {
		in   string
		want float64
	}{
		{"111K", float64(111000)},
		{"110T", float64(110000000000000)},
		{"120M", float64(120000000)},
		{"135G", float64(135000000000)},
		{"121", float64(121)},
		{"0", 0},
	}

	for _, c := range cases {
		got, err := SizeToBytes(c.in)
		if err != nil {
			t.Fatal(err)
		}
		if got != c.want {
			t.Errorf("SizeToBytes(%s) == %f, want %f", c.in, got, c.want)
		}
	}

	_, err := SizeToBytes("130X")
	if err == nil {
		t.Errorf("Should have received an error at undefined unit")
	}
}

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
