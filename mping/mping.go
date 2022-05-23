package mping

import (
	"bytes"
	cryptorand "crypto/rand"
	"encoding/binary"
	"log"
	"mumbleping/binaryhelper"
	"net"
	"strconv"
	"time"
)

func FormatVersion(v [4]byte) (s string) {
	for _, b := range v {
		if b != 0 {
			if s != "" {
				s += "."
			}
			s += strconv.FormatInt(int64(b), 10)
		}
	}
	return
}

type PingResult struct {
	Version          string        `json:"version"`
	UserC            uint32        `json:"user_c"`
	MaxUserC         uint32        `json:"max_user_c"`
	AllowedBandwidth uint32        `json:"allowed_bandwidth"`
	Latency          time.Duration `json:"latency"`
}

func Ping(l *log.Logger, ip string, timeout int64) (PingResult, error) {
	res := PingResult{}

	packetBuf := new(bytes.Buffer)

	packetBuf.Write([]byte{0x00, 0x00, 0x00, 0x00})

	identBuf := make([]byte, 8)
	if _, err := cryptorand.Read(identBuf); err != nil {
		return res, err
	}
	packetBuf.Write(identBuf)

	conn, err := net.Dial("udp", ip)
	if err != nil {
		return res, err
	}
	defer conn.Close()

	startTime := time.Now()

	if _, err := conn.Write(packetBuf.Bytes()); err != nil {
		return res, err
	}

read:

	const recvL = 24 // ping response packet is 24 bytes long

	recvBuf := make([]byte, recvL)

	conn.SetReadDeadline(time.Now().Add(time.Duration(time.Millisecond * time.Duration(timeout))))

	n, err := conn.Read(recvBuf)
	if err != nil {
		return res, err
	}
	if n < recvL {
		l.Printf("[ERROR] Short read! Read %d bytes but expected %d bytes", n, recvL)
		goto read
	}

	r := bytes.NewReader(recvBuf)

	var versionI [4]byte

	_, _ = r.Read(versionI[:])
	ident, _ := binaryhelper.ReadUint64(r)
	res.UserC, _ = binaryhelper.ReadUint32(r)
	res.MaxUserC, _ = binaryhelper.ReadUint32(r)
	res.AllowedBandwidth, _ = binaryhelper.ReadUint32(r)

	if ident != binary.BigEndian.Uint64(identBuf) {
		l.Printf("[ERROR] Wrong ident on ping recv, retrying...")
		goto read
	}

	endTime := time.Now()

	res.Latency = endTime.Sub(startTime)

	res.Version = FormatVersion(versionI)

	return res, nil
}
