package zconf

import (
	"github.com/gojekfarm/ziggurat-go/pkg/zbasic"
)

type DefaultValidator struct {
	rules map[string]func(config *zbasic.Config) error
}

func NewDefaultValidator(rules map[string]func(config *zbasic.Config) error) *DefaultValidator {
	return &DefaultValidator{rules: rules}
}

func (d *DefaultValidator) Validate(config *zbasic.Config) error {
	for _, ruleFunc := range d.rules {
		if err := ruleFunc(config); err != nil {
			return err
		}
	}
	return nil
}
