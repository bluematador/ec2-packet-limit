package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const RUNTIME = 24 * time.Hour
const POLL = time.Second
const AGGPOLL = time.Minute
const THREAD_MULTIPLE = 1
const PACKET_SIZE = 1
const SEND_LOOP = 10000
var iface string

func run(shutdownChannel chan struct{}) {
	conn, err := net.Dial("udp", "172.31.155.155:1000")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	data := []byte(strings.Repeat("a", PACKET_SIZE))

	for shutdown := false; !shutdown; {
		select {
		case <-shutdownChannel:
			shutdown = true
		default:
			for i := 0; i < SEND_LOOP; i += 1 {
				conn.Write(data)
			}
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

type aggregatedRate struct {
	min int
	max int
	sum int
	num int
}

func calc(shutdownChannel chan struct{}) {
	firstPackets := -1

	var lastPackets int
	var lastErr error

	rates := make([]int, RUNTIME / POLL)
	ratesIndex := 0

	aggRates := make([]aggregatedRate, RUNTIME / AGGPOLL)
	aggRatesIndex := 0

	aggMultiple := int(AGGPOLL) / int(POLL)

	rateTicker := time.Tick(POLL)
	for shutdown := false; !shutdown; {
		select {
		case <-shutdownChannel:
			shutdown = true
		case <-rateTicker:
			packets, err := getPackets()
			rate := -1

			if err != nil {
				fmt.Println("Error in stats", err)
			} else if firstPackets < 0 {
				firstPackets = packets
			} else if lastErr == nil {
				rate = packets - lastPackets
			}

			lastPackets = packets
			lastErr = err

			fmt.Println(ratesIndex, rate)
			rates[ratesIndex] = rate
			ratesIndex += 1

			if ratesIndex % aggMultiple == 0 {
				minuteMin := -1
				minuteMax := -1
				minuteSum := 0
				minuteNum := 0

				for _, r := range rates[(ratesIndex - aggMultiple) : ratesIndex] {
					if r >= 0 {
						minuteNum += 1
						minuteSum += r
						if r > minuteMax {
							minuteMax = r
						}
						if r < minuteMin || minuteMin < 0 {
							minuteMin = r
						}
					}
				}

				aggRates[aggRatesIndex] = aggregatedRate{minuteMin, minuteMax, minuteSum, minuteNum}
				fmt.Println(aggRatesIndex, aggRates[aggRatesIndex])
				aggRatesIndex += 1
			}

			if ratesIndex > len(rates) {
				shutdown = true
			}
		}
	}

	fmt.Println("save data to files")
}

func goWaitGroup(group *sync.WaitGroup, callback func()) {
	group.Add(1)
	go func() {
		defer group.Done()
		callback()
	}()
}

func handleShutdownSignals(shutdownChannel chan struct{}) {
	signal.Reset()

	first := true
	signals := []os.Signal{os.Interrupt, syscall.SIGTERM}
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, signals...)

	for shutdown := false; !shutdown; {
		select {
		case _, ok := <-sigchan:
			if !ok {
				fmt.Println("Unexpected channel close in signal handler. Shutting Down")
				close(shutdownChannel)
			} else {
				if first {
					fmt.Println("Signal Received. Cleaning Up. Signal again to forcefully exit.")
					first = false
					close(shutdownChannel)
				}
			}

		case <-shutdownChannel:
			shutdown = true
		}
	}

	signal.Reset()
}

func main() {
	shutdownChannel := make(chan struct{})
	go handleShutdownSignals(shutdownChannel)

	iface = findInterface()
	wg := &sync.WaitGroup{}

	goWaitGroup(wg, func() {
		calc(shutdownChannel)
	})
	for i := 0; i < runtime.NumCPU() * THREAD_MULTIPLE; i += 1 {
		goWaitGroup(wg, func() {
			run(shutdownChannel)
		})
	}

	wg.Wait()

	fmt.Println("Done!")
	os.Exit(0)
}
