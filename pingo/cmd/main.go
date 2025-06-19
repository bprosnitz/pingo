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

	for i := 0; i < n; i++ {
		select {
		case <-ctx.Done():
			return
		case <-time.After(sleep):
			break
		}
		sendPing(ctx, hostname, timeout)
	}
}

func sendPing(ctx context.Context, hostname string, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	elapsedTime, err := ping.Ping(ctx, hostname)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("elapsed time: %v\n", elapsedTime)
}
