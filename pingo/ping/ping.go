package ping

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/sync/errgroup"
)

func Ping(ctx context.Context, host string) (time.Duration, error) {
	addr, err := net.ResolveIPAddr("ip4", host)
	if err != nil {
		return 0, fmt.Errorf("error resolving host ip: %w", err)
	}

	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return 0, fmt.Errorf("error listening for packets: %w", err)
	}
	defer conn.Close()

	var outBuf [56]byte
	p, err := outgoingPayload(outBuf[:], rand.Int(), 0)
	if err != nil {
		return 0, fmt.Errorf("error creating outgoing payload: %w", err)
	}
	timeSent := time.Now()
	if _, err := conn.WriteTo(p, addr); err != nil {
		return 0, fmt.Errorf("error writing to connection: %w", err)
	}

	g, ctx := errgroup.WithContext(ctx)
	ctx, cancel := context.WithCancel(ctx)
	g.Go(func() error {
		<-ctx.Done()
		// End the read in the other goroutine.
		return conn.Close()
	})
	var timeElapsed time.Duration
	g.Go(func() error {
		// Stop the other goroutine on exit.
		defer cancel()
		var inBuf [56]byte
		_, _, err := conn.ReadFrom(inBuf[:])
		if err != nil {
			return fmt.Errorf("error reading from connection: %w", err)
		}
		timeElapsed = time.Since(timeSent)
		if err := checkIncomingPayload(inBuf[:]); err != nil {
			return fmt.Errorf("error: %w", err)
		}
		return nil
	})
	err = g.Wait()
	return timeElapsed, err
}

func outgoingPayload(buf []byte, id int, seq int) ([]byte, error) {
	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Body: &icmp.Echo{
			ID:   id,
			Seq:  seq,
			Data: nil,
		},
	}
	return msg.Marshal(buf)
}

func checkIncomingPayload(buf []byte) error {
	msg, err := icmp.ParseMessage(1, buf)
	if err != nil {
		return fmt.Errorf("error parsing mesage")
	}
	if msg.Body.Len(0) != 0 {
		return fmt.Errorf("expected 0 bytes in the body")
	}
	return nil
}
