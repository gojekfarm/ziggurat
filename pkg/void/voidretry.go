package void

import (
	"fmt"
	"github.com/gojekfarm/ziggurat-go/pkg/basic"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
)

type VoidRetry struct{}

func NewVoidRetry(c z.ConfigStore) z.MessageRetry {
	return &VoidRetry{}
}

func (v VoidRetry) Start(app z.App) error {
	return fmt.Errorf("error, no retry implementation found")
}

func (v VoidRetry) Retry(app z.App, payload basic.MessageEvent) error {
	return nil
}

func (v VoidRetry) Stop() error {
	return nil
}

func (v VoidRetry) Replay(app z.App, topicEntity string, count int) error {
	return fmt.Errorf("error, no retry implementation found")
}
