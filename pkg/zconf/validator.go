package zconf

import (
	"github.com/gojekfarm/ziggurat/pkg/zb"
)

type DefaultValidator struct {
	rules map[string]func(config *zb.Config) error
}

func NewDefaultValidator(rules map[string]func(config *zb.Config) error) *DefaultValidator {
	return &DefaultValidator{rules: rules}
}

func (d *DefaultValidator) Validate(config *zb.Config) error {
	for _, ruleFunc := range d.rules {
		if err := ruleFunc(config); err != nil {
			return err
		}
	}
	return nil
}
