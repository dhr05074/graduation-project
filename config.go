package main

import (
	"encoding/json"
	"fmt"
	xds "github.com/cncf/xds/go/xds/type/v3"
	"github.com/envoyproxy/envoy/contrib/golang/common/go/api"
	"github.com/envoyproxy/envoy/contrib/golang/filters/http/source/go/pkg/http"
	"google.golang.org/protobuf/types/known/anypb"
	"time"
	"traffic-shaping/manager"
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
}

var (
	bucketManager *manager.BucketManager
)

func ConfigFactory(c interface{}) api.StreamFilterFactory {
	conf, ok := c.(*config)
	if !ok {
		panic("unexpected config type")
	}

	if bucketManager == nil {
		bucketManager = manager.NewBucketManager(conf.Rate, 1*time.Millisecond)
		bucketManager.Start()
	}

	return func(callbacks api.FilterCallbackHandler) api.StreamFilter {
		return &filter{
			callbacks:     callbacks,
			config:        conf,
			bucketManager: bucketManager,
		}
	}
}
