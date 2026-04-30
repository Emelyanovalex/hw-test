package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	var timeout time.Duration
	flag.DurationVar(&timeout, "timeout", 10*time.Second, "connection timeout")
	flag.Parse()

	args := flag.Args()
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: go-telnet [--timeout=10s] host port")
		os.Exit(1)
	}

	address := net.JoinHostPort(args[0], args[1])
	client := NewTelnetClient(address, timeout, os.Stdin, os.Stdout)

	if err := client.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "...Connection error: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()
	fmt.Fprintf(os.Stderr, "...Connected to %s\n", address)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT)
	defer stop()

	sendDone := make(chan struct{})
	recvDone := make(chan struct{})

	go func() {
		defer close(sendDone)
		if err := client.Send(); err != nil {
			fmt.Fprintf(os.Stderr, "...Send error: %v\n", err)
			return
		}
		fmt.Fprintln(os.Stderr, "...EOF")
	}()

	go func() {
		defer close(recvDone)
		if err := client.Receive(); err != nil {
			fmt.Fprintf(os.Stderr, "...Receive error: %v\n", err)
			return
		}
		fmt.Fprintln(os.Stderr, "...Connection was closed by peer")
	}()

	select {
	case <-ctx.Done():
	case <-sendDone:
	case <-recvDone:
	}
}
