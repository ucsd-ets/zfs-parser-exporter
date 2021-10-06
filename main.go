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
	"time"
	"unicode"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

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
	if err != nil {
		return 0, fmt.Errorf("%v: sizeToBytes", err)
	}
	return sizeFloat, nil
}

func RunZpoolIOstat(zpoolCmd *string) (string, error) {
	out, err := exec.Command(*zpoolCmd, "iostat").Output()
	outS := string(out)
	if err != nil || outS == "" {
		return "No zpools detected", err
	}
	return outS, err
}

func Repeat(f func(), sleepTime int, stop chan int) {
	for {
		select {
		case <-stop:
			return
		default:
		}
		f()
		time.Sleep(time.Duration(sleepTime) * time.Second)
	}
}

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
	out, err := RunZpoolIOstat(zpoolCmd)
	if err != nil {
		log.Println("NO zpools detected. Are zpools available?")
		log.Fatal(err)
	}

	zpools, err := ParseZPoolIOStat(out, hostname)
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
			out, err := RunZpoolIOstat(zpoolCmd)
			log.Printf("%v", out)
			if err != nil {
				log.Fatal(err)
			}
			zpools, err := ParseZPoolIOStat(out, hostname)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Retrieved zpools %v", zpools)

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
