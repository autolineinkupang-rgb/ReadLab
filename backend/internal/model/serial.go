package model

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"sync/atomic"
	"time"
)

var serialCounter uint64

func NewSerial() string {
	b := make([]byte, 16)
	rand.Read(b)
	n := atomic.AddUint64(&serialCounter, 1)
	h := sha256.Sum256([]byte(fmt.Sprintf("%d-%d-%x-%d", time.Now().UnixNano(), n, b, n)))
	return fmt.Sprintf("%x", h[:16])
}
