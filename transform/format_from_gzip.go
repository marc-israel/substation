package transform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brexhq/substation/config"
	"github.com/brexhq/substation/message"
)

func newFormatFromGzip(_ context.Context, cfg config.Config) (*formatFromGzip, error) {
	conf := formatGzipConfig{}
	if err := conf.Decode(cfg.Settings); err != nil {
		return nil, fmt.Errorf("transform: format_from_gzip: %v", err)
	}

	tf := formatFromGzip{
		conf:     conf,
		isObject: false,
	}

	return &tf, nil
}

type formatFromGzip struct {
	conf     formatGzipConfig
	isObject bool
}

func (tf *formatFromGzip) Transform(ctx context.Context, msg *message.Message) ([]*message.Message, error) {
	if msg.IsControl() {
		return []*message.Message{msg}, nil
	}

	gz, err := fmtFromGzip(msg.Data())
	if err != nil {
		return nil, fmt.Errorf("transform: format_from_gzip: %v", err)
	}

	msg.SetData(gz)
	return []*message.Message{msg}, nil
}

func (tf *formatFromGzip) String() string {
	b, _ := json.Marshal(tf.conf)
	return string(b)
}
