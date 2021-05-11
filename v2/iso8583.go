package iso8583v2

import (
	"sync"
	"sync/atomic"
)

var (
	tagCache      map[string]atomic.Value
	fixedTagCache map[string]atomic.Value

	tagLock      sync.RWMutex
	fixedTagLock sync.RWMutex
)

func init() {
	tagCache = make(map[string]atomic.Value)
	fixedTagCache = make(map[string]atomic.Value)
}
