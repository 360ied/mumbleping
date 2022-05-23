package binaryhelper

import (
	"bytes"
	"encoding/binary"
)

func ReadUint32(r *bytes.Reader) (uint32, error) {
	b := make([]byte, 4)
	if _, err := r.Read(b); err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint32(b), nil
}

func ReadUint64(r *bytes.Reader) (uint64, error) {
	b := make([]byte, 8)
	if _, err := r.Read(b); err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint64(b), nil
}
