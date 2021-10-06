package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"
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

func fakeFunc() {
	seed := rand.NewSource(time.Now().UnixNano())
	randGen := rand.New(seed)
	randFloat := randGen.Float32()
	fmt.Println(randFloat)
}

func TestRepeat(t *testing.T) {
	rescueStdout := os.Stdout
	c := make(chan int)
	r, w, _ := os.Pipe()
	os.Stdout = w

	go Repeat(fakeFunc, 1, c)
	for i := 0; i < 3; i++ {
		time.Sleep(time.Duration(1) * time.Second)
		if i == 2 {
			c <- 1
			w.Close()
		}
	}
	out, _ := ioutil.ReadAll(r)
	os.Stdout = rescueStdout
	nums := strings.Split(string(out), "\n")

	// verify that the set of nums is unique
	numsSeen := make(map[string]bool)
	for _, num := range nums {
		if _, entry := numsSeen[num]; !entry {
			numsSeen[num] = true
		} else {
			t.Fatal("repeat didnt recall")
		}
	}
}
