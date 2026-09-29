package main

import (
	"context"
	"errors"
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

type telnetClient struct {
	ctx     context.Context
	address string
	in      io.Reader
	out     io.Writer
	conn    net.Conn
}

func NewTelnetClient(
	ctx context.Context,
	host string,
	port string,
	_ time.Duration,
	in io.ReadCloser,
	out io.Writer,
) (TelnetClient, error) {
	if host == "" || port == "" {
		return nil, errors.New("host and port should be defined")
	}

	address := net.JoinHostPort(host, port)
	if _, _, err := net.SplitHostPort(host); err == nil {
		address = host
	}

	client := &telnetClient{
		ctx:     ctx,
		address: address,
		in:      in,
		out:     out,
	}
	return client, nil
}

func (t *telnetClient) Connect() error {
	dialer := net.Dialer{}
	conn, err := dialer.DialContext(t.ctx, "tcp", t.address)
	if err != nil {
		return fmt.Errorf("can't connect by host:port cause: %w", err)
	}
	t.conn = conn
	return nil
}

func (t *telnetClient) Close() error {
	if t.conn == nil {
		return nil
	}
	return t.conn.Close()
}

func (t *telnetClient) Send() error {
	_, err := io.Copy(t.conn, t.in)
	if err != nil {
		return fmt.Errorf("error while sending: %w", err)
	}
	return nil
}

func (t *telnetClient) Receive() error {
	_, err := io.Copy(t.out, t.conn)
	if err != nil {
		return fmt.Errorf("error while receiving: %w", err)
	}
	return nil
}

// Place your code here.
// P.S. Author's solution takes no more than 50 lines.
