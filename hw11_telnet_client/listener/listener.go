// Package main implements a simple TCP listener for manual TELNET testing.
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Printf("INITIALIZING CONNECTION\n")

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "quit" || text == "exit" {
			break
		}
		log.Println(text)
		if _, err := fmt.Fprintf(conn, "Received data from server: '%s'\n", text); err != nil {
			log.Printf("Cannot write to connection: %v", err)
			return
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error happened on connection with %s: %v", conn.RemoteAddr(), err)
	}

	log.Printf("Closing connection with %s", conn.RemoteAddr())
}

func main() {
	lc := net.ListenConfig{}
	l, err := lc.Listen(context.Background(), "tcp", "127.0.0.1:8081")
	if err != nil {
		log.Fatalf("Cannot listen: %v", err)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("Cannot accept: %v", err)
			return
		}

		go handleConnection(conn)
	}
}
