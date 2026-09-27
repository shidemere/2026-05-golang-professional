// Package main implements a primitive TELNET client.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	host, port, err := GetHostAndPort()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error occurred: %v\n", err)
		os.Exit(1)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	client := NewTelnetClient(ctx, host, port, os.Stdin, os.Stdout)
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

func GetHostAndPort() (string, string, error) {
	args := os.Args[1:]
	if len(args) != 2 {
		return "", "", fmt.Errorf("can't parse CLI argument: need to define host & port")
	}
	return args[0], args[1], nil
}
