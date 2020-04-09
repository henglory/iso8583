package iso8583v2

import "sync/atomic"

var (
	tagCache      map[string]atomic.Value
	fixedTagCache map[string]atomic.Value
)

func init() {
	tagCache = make(map[string]atomic.Value)
	fixedTagCache = make(map[string]atomic.Value)
}
