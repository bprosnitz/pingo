package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bprosnitz/pingo/pingo/ping"
)

func main() {
	var timeout time.Duration
	flag.DurationVar(&timeout, "timeout", 10*time.Second, "Amount of time to wait before cancelling the ping request")
	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatalf("expected target hostname arg")
	}
	hostname := flag.Arg(0)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, timeout)
	defer cancel()

	elapsedTime, err := ping.Ping(ctx, hostname)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("elapsed time: %v\n", elapsedTime)
}
