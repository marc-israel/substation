package transform

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"

	"github.com/brexhq/substation/config"
	iconfig "github.com/brexhq/substation/internal/config"
	"github.com/brexhq/substation/internal/metrics"
	"github.com/brexhq/substation/message"
)

type utilityMetricsCountConfig struct {
	Metric iconfig.Metric `json:"metric"`
}

func (c *utilityMetricsCountConfig) Decode(in interface{}) error {
	return iconfig.Decode(in, c)
}

func newUtilityMetricCount(ctx context.Context, cfg config.Config) (*utilityMetricsCount, error) {
	// conf gets validated when calling metrics.New.
	conf := utilityMetricsCountConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: utility_metric_count: %v", err)
	}

	m, err := metrics.New(ctx, conf.Metric.Destination)
	if err != nil {
		return nil, fmt.Errorf("transform: utility_metric_count: %v", err)
	}

	tf := utilityMetricsCount{
		conf:   conf,
		metric: m,
	}

	return &tf, nil
}

type utilityMetricsCount struct {
	conf utilityMetricsCountConfig

	metric metrics.Generator
	count  uint32
}

func (tf *utilityMetricsCount) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		if err := tf.metric.Generate(ctx, metrics.Data{
			Name:       tf.conf.Metric.Name,
			Value:      tf.count,
			Attributes: tf.conf.Metric.Attributes,
		}); err != nil {
			return nil, fmt.Errorf("transform: utility_metric_count: %v", err)
		}

		atomic.StoreUint32(&tf.count, 0)
		return []*message.Message{msg}, nil
	}

	atomic.AddUint32(&tf.count, 1)
	return []*message.Message{msg}, nil
}

func (tf *utilityMetricsCount) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
