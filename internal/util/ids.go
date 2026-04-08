package util

import (
	"fmt"
	"sync/atomic"
	"time"
)

var idCounter uint64

func NewID(prefix string) string {
	n := atomic.AddUint64(&idCounter, 1)
	return fmt.Sprintf("%s_%d_%d", prefix, time.Now().UTC().UnixNano(), n)
}
