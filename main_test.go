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

func TestFetchZPools(t *testing.T) {
	hostname := hostname(t)
	zpools, err := FetchZPools("/sbin/zpool", hostname)
	if err != nil {
		t.Fatal(err)
	}
	zpoolsType := fmt.Sprintf("%T", zpools)
	if zpoolsType != "[]main.ZPool" {
		t.Errorf("Should have been type []main.Zpool, not %s", zpoolsType)
	}

}

func TestRegister(t *testing.T) {
	hostname := hostname(t)
	zpools, err := FetchZPools("/sbin/zpool", hostname)
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
