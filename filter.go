package main

import (
	"github.com/envoyproxy/envoy/contrib/golang/common/go/api"
	"strconv"
	"traffic-shaping/manager"
)

type filter struct {
	api.PassThroughStreamFilter

	callbacks api.FilterCallbackHandler
	config    *config

	bucketManager *manager.BucketManager

	reqChan map[int]chan struct{}
}

func (f *filter) DecodeHeaders(header api.RequestHeaderMap, endStream bool) api.StatusType {
	priorityHeader, ok := header.Get("X-Priority")
	if !ok {
		priorityHeader = "5"
	}

	// priority value must be between 1 and 5
	priorityValue, err := strconv.Atoi(priorityHeader)
	if err != nil || priorityValue < 1 || priorityValue > 5 {
		f.callbacks.SendLocalReply(500, "Invalid priority value. It must be between 1 and 5.", nil, 0, "")
	}

	if err := f.bucketManager.Consume(uint(priorityValue)); err != nil {
		f.callbacks.SendLocalReply(429, "Too Many Requests. Try again later.", nil, 0, "")
	}

	return api.Continue
}

func main() {
}
