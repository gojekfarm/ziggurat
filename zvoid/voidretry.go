package zvoid

import (
	"fmt"
	"github.com/gojekfarm/ziggurat/zbase"
	"github.com/gojekfarm/ziggurat/ztype"
)

type VoidRetry struct{}

func NewRetry() ztype.MessageRetry {
	return &VoidRetry{}
}

func (v VoidRetry) Start(app ztype.App) error {
	return fmt.Errorf("error, no retry implementation found")
}

func (v VoidRetry) Retry(app ztype.App, payload zbase.MessageEvent) error {
	return nil
}

func (v VoidRetry) Stop(a ztype.App) {

}

func (v VoidRetry) Replay(app ztype.App, topicEntity string, count int) error {
	return fmt.Errorf("error, no retry implementation found")
}
