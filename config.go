package main

import (
	"encoding/json"
	"fmt"
	"github.com/alitto/pond"
	xds "github.com/cncf/xds/go/xds/type/v3"
	"github.com/envoyproxy/envoy/contrib/golang/common/go/api"
	"github.com/envoyproxy/envoy/contrib/golang/filters/http/source/go/pkg/http"
	"golang.org/x/time/rate"
	"google.golang.org/protobuf/types/known/anypb"
	"sync"
	"time"
)

type config struct {
	Paths   map[string]int `json:"paths"`
	Workers int            `json:"workers"`
	Rate    int            `json:"rate"`
	Burst   int            `json:"burst"`
}

type parser struct {
}

func (p *parser) Parse(any *anypb.Any) (interface{}, error) {
	var confStruct xds.TypedStruct
	if err := any.UnmarshalTo(&confStruct); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	var conf config
	paths := confStruct.GetValue().AsMap()

	pathsJson, _ := json.Marshal(paths)
	if err := json.Unmarshal(pathsJson, &conf); err != nil {
		return &conf, fmt.Errorf("failed to unmarshal paths: %w", err)
	}

	return &conf, nil
}

func (p *parser) Merge(parentConfig interface{}, childConfig interface{}) interface{} {
	return childConfig
}

func init() {
	http.RegisterHttpFilterConfigFactoryAndParser("traffic", ConfigFactory, &parser{})

	go func() {
		for {
			select {
			case <-time.Tick(1 * time.Second):
				fmt.Printf("cntMap: %+v\n", cntMap)
			case p := <-cntChan:
				cntMap[p] += 1
			}
		}
	}()

	for i := 0; i < 5; i++ {
		go func() {
			for {
				<-processChan
				time.Sleep(1 * time.Millisecond)
			}
		}()
	}
}

var (
	cntMap      = make(map[int]int)
	pm          = make(map[int]*pond.WorkerPool)
	mu          = &sync.Mutex{}
	cntChan     = make(chan int, 1000)
	limiter     *rate.Limiter
	reqChan     = make(map[int]chan struct{})
	processChan = make(chan struct{})
)

func ConfigFactory(c interface{}) api.StreamFilterFactory {
	conf, ok := c.(*config)
	if !ok {
		panic("unexpected config type")
	}

	if limiter == nil {
		limiter = rate.NewLimiter(rate.Every(time.Duration(conf.Rate)*time.Millisecond), 2<<conf.Burst)
	}

	return func(callbacks api.FilterCallbackHandler) api.StreamFilter {
		return &filter{
			callbacks: callbacks,
			config:    conf,
			cntMap:    cntMap,
			mu:        mu,
			pm:        pm,
			lim:       limiter,
			reqChan:   reqChan,
		}
	}
}
