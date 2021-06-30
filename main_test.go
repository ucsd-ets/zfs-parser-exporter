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
		want int
	}{
		{"111K", 111000},
		{"110T", 110000000000000},
		{"120M", 120000000},
		{"135G", 135000000000},
		{"121", 121},
		{"0", 0},
	}

	for _, c := range cases {
		got, err := SizeToBytes(c.in)
		if err != nil {
			t.Fatal(err)
		}
		if got != c.want {
			fmt.Println(got)
			t.Errorf("SizeToBytes(%q) == %q, want %q", c.in, got, c.want)
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
	gr := NewGaugeRegistry()
	gr.PrometheusRegistry = prometheus.NewRegistry()

	gr.Register(zpools[0])

	if len(gr.gauges) == 0 {
		t.Error("No gauges were registered")
	}
}
