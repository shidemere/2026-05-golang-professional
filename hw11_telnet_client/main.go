// Package main implements a primitive TELNET client.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	host, port, timeout, err := GetHostAndPort()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error occurred: %v\n", err)
		os.Exit(1)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	client, err := NewTelnetClient(ctx, host, port, timeout, os.Stdin, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't create client: %v\n", err)
		os.Exit(1)
	}
	err = client.Connect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't connect: %v\n", err)
		os.Exit(1)
	}
	defer cancel()
	fmt.Fprintf(os.Stderr, "...Connected to %s:%s\n", host, port)

	go func() {
		if err := client.Send(); err != nil {
			fmt.Fprintln(os.Stderr, err)
		} else {
			fmt.Fprintln(os.Stderr, "...EOF")
		}
		cancel()
	}()

	go func() {
		if err := client.Receive(); err != nil {
			if ctx.Err() == nil {
				fmt.Fprintln(os.Stderr, err)
			}
		} else {
			fmt.Fprintln(os.Stderr, "...Connection was closed by peer")
		}
		cancel()
	}()

	<-ctx.Done()
	err = client.Close()
	if err != nil && ctx.Err() == nil {
		fmt.Fprintf(os.Stderr, "error while closing connection: %v\n", err)
	}
}

func GetHostAndPort() (string, string, time.Duration, error) {
	var timeout time.Duration
	flag.DurationVar(&timeout, "timeout", 10*time.Second, "connection timeout")
	flag.Parse()

	args := flag.Args()
	if len(args) != 2 {
		return "", "", 0, fmt.Errorf("can't parse CLI argument: need to define host & port")
	}
	return args[0], args[1], timeout, nil
}
