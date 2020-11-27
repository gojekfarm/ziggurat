package void

import (
	"fmt"
	"github.com/gojekfarm/ziggurat-go/pkg/z"
	"github.com/gojekfarm/ziggurat-go/pkg/zbasic"
)

type VoidRetry struct{}

func NewRetry(c z.ConfigStore) z.MessageRetry {
	return &VoidRetry{}
}

func (v VoidRetry) Start(app z.App) error {
	return fmt.Errorf("error, no retry implementation found")
}

func (v VoidRetry) Retry(app z.App, payload zbasic.MessageEvent) error {
	return nil
}

func (v VoidRetry) Stop(a z.App) {

}

func (v VoidRetry) Replay(app z.App, topicEntity string, count int) error {
	return fmt.Errorf("error, no retry implementation found")
}
