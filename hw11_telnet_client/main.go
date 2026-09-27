package main

import (
	"io"
	"log"
	"os"
	"time"
)

func main() {
	rc := io.NopCloser(os.Stdin)

	_, err := NewTelnetClient("localhost", "8081", time.Hour, rc, os.Stdout)
	if err != nil {
		log.Fatal("Цхьа бид бы хьу")
	}
}
