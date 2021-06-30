package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Facade around registering & updating exports
type GaugeRegistry struct {
	gauges             map[string]prometheus.Gauge
	Hostname           string
	PrometheusRegistry *prometheus.Registry
	Namespace          string
}

func NewGaugeRegistry(prometheusNamespace string) *GaugeRegistry {
	gr := GaugeRegistry{}
	gr.gauges = make(map[string]prometheus.Gauge)
	gr.Namespace = prometheusNamespace
	return &gr
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

func (gr *GaugeRegistry) Update(z ZPool) {
	gr.gauges[z.CapAlloc.Name].Set(z.CapAlloc.Val)
	gr.gauges[z.CapFree.Name].Set(z.CapFree.Val)
	gr.gauges[z.OpsRead.Name].Set(z.OpsRead.Val)
	gr.gauges[z.OpsWrite.Name].Set(z.OpsWrite.Val)
	gr.gauges[z.BdwRead.Name].Set(z.BdwRead.Val)
	gr.gauges[z.BdwWrite.Name].Set(z.BdwWrite.Val)
}

func (gr *GaugeRegistry) registerGauge(s Stat) {
	g := promauto.NewGauge(prometheus.GaugeOpts{
		Name:        s.Name,
		Namespace:   gr.Namespace,
		Help:        s.Help,
		ConstLabels: s.Labels,
	})
	gr.gauges[s.Name] = g

	g.Set(s.Val)
	gr.PrometheusRegistry.MustRegister(g)
}

type Stat struct {
	Name   string
	Help   string
	Val    float64
	Labels map[string]string
}

type ZPool struct {
	Name     string
	CapAlloc Stat
	CapFree  Stat
	OpsRead  Stat
	OpsWrite Stat
	BdwRead  Stat
	BdwWrite Stat
}

/** UTIL FUNCTIONS **/
// https://stackoverflow.com/questions/38554353/how-to-check-if-a-string-only-contains-alphabetic-characters-in-go
func IsLetter(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func SizeToBytes(size string) (float64, error) {
	// base case, if single alphanumeric
	if len(size) == 1 {
		val, err := strconv.ParseFloat(size, 2)
		if err != nil {
			return 0, fmt.Errorf("%v: sizeToBytes", err)
		}
		return val, nil
	}

	lastVal := string(size[len(size)-1:])
	unconvertedVal, err := strconv.ParseFloat(size[0:len(size)-1], 2)

	if err != nil {
		return 0, fmt.Errorf("%v: sizeToBytes", err)
	}

	switch lastVal {
	case "K":
		return unconvertedVal * 1000, nil
	case "M":
		return unconvertedVal * 1000000, nil
	case "G":
		return unconvertedVal * 1000000000, nil
	case "T":
		return unconvertedVal * 1000000000000, nil
	}

	// uncaught case
	if IsLetter(lastVal) {
		return 0, errors.New(fmt.Sprintf("Uncaught unit type %s", lastVal))
	}

	// no units
	sizeFloat, err := strconv.ParseFloat(size, 2)
	return sizeFloat, nil
}

func ParseZPoolIOStat(zpoolOutput string, hostname string) ([]ZPool, error) {
	zpools := []ZPool{}
	splitStatsTbl := strings.Split(zpoolOutput, "\n")
	trimLength := 2
	if len(splitStatsTbl) == 5 {
		trimLength = 1
	}
	// first 3 are headers, last row is just "---"
	for _, s := range splitStatsTbl[3 : len(splitStatsTbl)-trimLength] {
		// fields are name, capacity_alloc, capacity_free, ops_read, ops_write, bdw_read, bdw_write
		fields := strings.Fields(s)
		labels := map[string]string{"zpool_name": fields[0], "hostname": hostname}

		capAlloc, err := SizeToBytes(fields[1])
		if err != nil {
			return zpools, fmt.Errorf("%v: Could not convert capAlloc field", err)
		}
		capAllocStats := Stat{
			"zpool_capacity_allocable_bytes",
			"Free capacity allocable in bytes",
			float64(capAlloc),
			labels,
		}

		capFree, err := SizeToBytes(fields[2])
		if err != nil {
			return zpools, fmt.Errorf("%v: Could not convert capAlloc field", err)
		}
		capFreeStats := Stat{
			"zpool_capacity_free_bytes",
			"Free capacity space in bytes",
			float64(capFree),
			labels,
		}

		opRead, err := SizeToBytes(fields[3])
		if err != nil {
			return zpools, fmt.Errorf("%v: Could not convert capAlloc field", err)
		}
		opReadStats := Stat{
			"zpool_operations_read_bytes",
			"zpool operations read in bytes",
			float64(opRead),
			labels,
		}

		opWrite, err := SizeToBytes(fields[4])
		if err != nil {
			return zpools, fmt.Errorf("%v: Could not convert capAlloc field", err)
		}
		opWriteStats := Stat{
			"zpool_operations_write_bytes",
			"zpool operations write in bytes",
			float64(opWrite),
			labels,
		}

		bdwRead, err := SizeToBytes(fields[5])
		if err != nil {
			return zpools, fmt.Errorf("%v: Could not convert capAlloc field", err)
		}
		bdwReadStats := Stat{
			"zpool_bandwidth_read_bytes",
			"zpool bandwidth read in bytes",
			float64(bdwRead),
			labels,
		}

		bdwWrite, err := SizeToBytes(fields[6])
		if err != nil {
			return zpools, fmt.Errorf("%v: Could not convert capAlloc field", err)
		}
		bdwWriteStats := Stat{
			"zpool_bandwidth_write_bytes",
			"zpool bandwidth write in bytes",
			float64(bdwWrite),
			labels,
		}

		zpool := ZPool{
			fields[0],
			capAllocStats,
			capFreeStats,
			opReadStats,
			opWriteStats,
			bdwReadStats,
			bdwWriteStats,
		}

		zpools = append(zpools, zpool)
	}

	return zpools, nil
}

/** /UTIL FUNCTIONS **/

func main() {
	runningPort := flag.Int("port", 2112, "Host port for zfs exporter service. Defaults to 2112")
	zpoolCmd := flag.String("zpool-path", "/sbin/zpool", "Path for zpool command. Defaults to /sbin/zpool")
	parseSeconds := flag.Int("parse-seconds", 2, "Seconds to wait before rerunning zpool command")
	prometheusNamespace := flag.String("namespace", "", "Namespace (i.e. a prefix) for exported Prometheus timeseries. Defaults to ''")

	flag.Parse()

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	// FIXME move running command to external and pass output into input of this function
	out, err := exec.Command(*zpoolCmd, "iostat").Output()
	if err != nil {
		log.Fatal(err)
	}
	outS := string(out)
	zpools, err := ParseZPoolIOStat(outS, hostname)
	if err != nil {
		log.Fatal(err)
	}

	r := prometheus.NewRegistry()
	gr := NewGaugeRegistry(*prometheusNamespace)
	gr.PrometheusRegistry = r

	for _, zpool := range zpools {
		gr.Register(zpool)
	}

	go func() {
		for {
			out, err := exec.Command(*zpoolCmd, "iostat").Output()
			if err != nil {
				log.Fatal(err)
			}
			outS := string(out)
			zpools, err := ParseZPoolIOStat(outS, hostname)
			if err != nil {
				log.Fatal(err)
			}
			for _, zpool := range zpools {
				gr.Update(zpool)
			}
			time.Sleep(time.Duration(*parseSeconds) * time.Second)
		}
	}()

	handler := promhttp.HandlerFor(r, promhttp.HandlerOpts{})
	listenAddress := ":" + strconv.Itoa(*runningPort)
	log.Printf("Listening on %s. Running commands every %d seconds", listenAddress, *parseSeconds)

	http.Handle("/metrics", handler)
	http.ListenAndServe(listenAddress, nil)
}
