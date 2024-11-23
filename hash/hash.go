package hash

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"time"
)

func Now() [32]byte {
	n := time.Now().Unix()
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(n))
	return sha256.Sum256(b)
}

func SNow() string {
	h := Now()
	s := ""
	for i := range h {
		s += fmt.Sprintf("%03d", h[i])
	}
	return s
}
