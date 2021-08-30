package main

import (
	"bytes"
	cryptorand "crypto/rand"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"time"
)

func main() {
	var timeout int64
	flag.Int64Var(&timeout, "timeout", 1000, "Ping timeout in milliseconds")

	flag.Parse()

	ip := flag.Arg(0)
	if ip == "" {
		log.Fatalf("Args Error: IP Address not provided.")
	}

	packetBuf := &bytes.Buffer{}

	packetBuf.Write([]byte{0x00, 0x00, 0x00, 0x00})

	identBuf := make([]byte, 8)
	if _, err := cryptorand.Read(identBuf); err != nil {
		panic(err)
	}
	packetBuf.Write(identBuf)

	conn, err := net.Dial("udp", ip)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	if _, err := conn.Write(packetBuf.Bytes()); err != nil {
		panic(err)
	}

read:

	const recvL = 24 // ping response packet is 24 bytes long

	recvBuf := make([]byte, recvL)

	conn.SetReadDeadline(time.Now().Add(time.Duration(time.Millisecond * time.Duration(timeout))))

	n, err := conn.Read(recvBuf)
	if err != nil {
		panic(err)
	}
	if n < recvL {
		log.Printf("[ERROR] Short read! Read %d bytes but expected %d bytes", n, recvL)
		goto read
	}

	r := bytes.NewReader(recvBuf)

	version, _ := ReadUint32(r)
	ident, _ := ReadUint64(r)
	userC, _ := ReadUint32(r)
	maxUserC, _ := ReadUint32(r)
	allowedBandwith, _ := ReadUint32(r)

	if ident != binary.BigEndian.Uint64(identBuf) {
		log.Printf("[ERROR] Wrong ident on ping recv, retrying...")
		goto read
	}

	fmt.Printf("-----Ping-----\nVersion: %d\nUsers: %d/%d\nAllowed Bandwidth: %d\n--------------\n", version, userC, maxUserC, allowedBandwith)
}
