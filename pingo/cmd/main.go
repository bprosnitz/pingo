package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bprosnitz/pingo/pingo/ping"
)

func main() {
	var timeout time.Duration
	var sleep time.Duration
	var n int
	flag.DurationVar(&timeout, "timeout", 10*time.Second, "Amount of time to wait before cancelling the ping request")
	flag.DurationVar(&sleep, "sleep", time.Second, "Amount of time to wait for the next ping")
	flag.IntVar(&n, "n", math.MaxInt, "Number of pings to send")
	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatalf("expected target hostname arg")
	}
	hostname := flag.Arg(0)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var (
		i             int
		numErrors     int
		totalDuration time.Duration
	)
	defer func() {
		fmt.Printf("Summary:\n")
		fmt.Printf("%d requests\n", i)
		fmt.Printf("%d errors (%f%%)\n", numErrors, float64(numErrors)/float64(i))
		fmt.Printf("average latency: %v\n", totalDuration/time.Duration(i))
	}()
	for ; i < n; i++ {
		select {
		case <-ctx.Done():
			return
		case <-time.After(sleep):
			break
		}
		elapsed, err := sendPing(ctx, hostname, timeout)
		if err != nil {
			fmt.Printf("error: %v", err)
			numErrors++
			continue
		}
		totalDuration += elapsed
		fmt.Printf("elapsed time: %v\n", elapsed)
	}
}

func sendPing(ctx context.Context, hostname string, timeout time.Duration) (time.Duration, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return ping.Ping(ctx, hostname)
}
