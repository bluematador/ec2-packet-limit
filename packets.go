package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const RUNTIME = 24 * time.Hour
const POLL = 1 * time.Second
const THREAD_MULTIPLE = 1
var iface string

func run(i int) {
	fmt.Println("Running", i)
	conn, err := net.Dial("udp", "172.31.155.155:1000")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	for {
		conn.Write([]byte("a"))
	}
}

func pretty(n int) string {
	in := strconv.FormatInt(int64(n), 10)
	out := make([]byte, len(in)+(len(in)-2+int(in[0]/'0'))/3)
	if in[0] == '-' {
		in, out[0] = in[1:], '-'
	}

	for i, j, k := len(in)-1, len(out)-1, 0; ; i, j = i-1, j-1 {
		out[j] = in[i]
		if i == 0 {
			return string(out)
		}
		if k++; k == 3 {
			j, k = j-1, 0
			out[j] = ','
		}
	}
}

func findInterface() string {
	files, err := ioutil.ReadDir("/sys/class/net/")
	if err != nil {
		panic(err)
	}

	for _, f := range files {
		if f.Name() != "lo" {
			return f.Name()
		}
	}

	panic("no default interface!")
}

func getPackets() (int, error) {
	data, err := ioutil.ReadFile("/sys/class/net/" + iface + "/statistics/tx_packets")
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(data)))
}

func calc() {
	firstPackets := -1
	maxPackets := -1
	firstStart := time.Now()
	var lastPackets int
	var lastErr error

	for {
		packets, err := getPackets()
		// if err != nil {
		// 	panic(err)
		// }
		if err == nil && firstPackets < 0 {
			firstPackets = packets
			firstStart = time.Now()
		} else if err == nil && lastErr == nil {
			rate := packets - lastPackets
			if rate > maxPackets {
				maxPackets = rate
			}

			avg := int(float64(packets - firstPackets) / time.Now().Sub(firstStart).Seconds())

			fmt.Println(time.Now().Format("15:04:05"), "Rate:", pretty(rate), POLL, "Avg:", pretty(avg), "1s", "Max:", pretty(maxPackets), POLL)
		}

		time.Sleep(POLL)
		lastPackets = packets
		lastErr = err
	}
}

func main() {
	iface = findInterface()

	go calc()
	for i := 0; i < runtime.NumCPU() * THREAD_MULTIPLE; i++ {
		go run(i)
	}

	time.Sleep(RUNTIME)

	fmt.Println("Done!")
	os.Exit(0)
}
