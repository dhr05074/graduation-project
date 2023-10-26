package main

import (
	"github.com/alitto/pond"
	"github.com/envoyproxy/envoy/contrib/golang/common/go/api"
	"golang.org/x/time/rate"
	"sync"
	"time"
)

type filter struct {
	api.PassThroughStreamFilter

	callbacks api.FilterCallbackHandler
	config    *config
	pm        map[int]*pond.WorkerPool
	cntMap    map[int]int
	mu        *sync.Mutex
	lim       *rate.Limiter

	reqChan map[int]chan struct{}
}

func (f *filter) DecodeHeaders(header api.RequestHeaderMap, endStream bool) api.StatusType {
	pt, _ := header.Get(":path")

	p, ok := f.config.Paths[pt]
	if !ok {
		p = 50
	}

	go func() {
		defer f.callbacks.RecoverPanic()

		time.Sleep(1 * time.Millisecond)

		f.mu.Lock()
		if _, ok := f.reqChan[p]; !ok {
			f.reqChan[p] = make(chan struct{}, 1<<10>>uint(p))
		}
		f.mu.Unlock()

		select {
		case f.reqChan[p] <- struct{}{}:
			f.callbacks.Continue(api.Continue)
		default:
			f.callbacks.SendLocalReply(429, "Too Many Requests", nil, 0, "")
		}
	}()

	return api.Running
}

func main() {
}
