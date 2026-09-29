#!/usr/bin/env bash
set -xeuo pipefail

SERVER_GO=""
function cleanup() {
  rm -f go-telnet
  if [ -n "${SERVER_GO}" ]; then
    rm -f "${SERVER_GO}"
  fi
}
trap cleanup EXIT

go build -o go-telnet

SERVER_GO=$(mktemp /tmp/go-telnet-server-XXXXXX.go)
cat >"${SERVER_GO}" <<'EOF'
package main

import (
	"io"
	"log"
	"net"
	"os"
)

func main() {
	l, err := net.Listen("tcp", "localhost:4242")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	conn, err := l.Accept()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	if _, err = conn.Write([]byte("Hello\nFrom\nNC\n")); err != nil {
		log.Fatal(err)
	}

	out, err := os.Create("/tmp/nc.out")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	if _, err = io.Copy(out, conn); err != nil {
		log.Fatal(err)
	}
}
EOF

go run "${SERVER_GO}" &
NC_PID=$!

sleep 1
(echo -e "I\nam\nTELNET client\n" && cat 2>/dev/null) | ./go-telnet --timeout=5s localhost 4242 >/tmp/telnet.out &
TL_PID=$!

sleep 5
kill ${TL_PID} 2>/dev/null || true
kill ${NC_PID} 2>/dev/null || true

function fileEquals() {
  local fileData
  fileData=$(cat "$1")
  [ "${fileData}" = "${2}" ] || (echo -e "unexpected output, $1:\n${fileData}" && exit 1)
}

expected_nc_out='I
am
TELNET client'
fileEquals /tmp/nc.out "${expected_nc_out}"

expected_telnet_out='Hello
From
NC'
fileEquals /tmp/telnet.out "${expected_telnet_out}"

echo "PASS"
