package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"time"
)

type TelnetClient interface {
	Connect() error
	io.Closer
	Send() error
	Receive() error
}

func NewTelnetClient(host string, port string, timeout time.Duration, in io.ReadCloser, out io.Writer) (TelnetClient, error) {
	defer in.Close()
	ip, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		return nil, fmt.Errorf("can't resolve host ip: %v", err)
	}

	address := net.JoinHostPort(ip.String(), port)
	dialer := net.Dialer{}

	conn, err := dialer.DialContext(context.Background(), "tcp", address)
	if err != nil {
		return nil, fmt.Errorf("can't connect by host:port cause: %v", err)
	}
	defer conn.Close()

	// TODO: Need to extract in separate gorutine
	go WriteToClient(in, conn)
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		text := scanner.Text()
		out.Write([]byte(text))
	}
	return nil, nil
}

func WriteToClient(in io.ReadCloser, conn net.Conn) {
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		text := scanner.Text()
		text = text + "\n"
		// fmt.Fprint(os.Stderr, text)
		conn.Write([]byte(text))
	}
}

// Place your code here.
// P.S. Author's solution takes no more than 50 lines.
