package zconf

import (
	"github.com/gojekfarm/ziggurat/zbase"
)

type DefaultValidator struct {
	rules map[string]func(config *zbase.Config) error
}

func NewDefaultValidator(rules map[string]func(config *zbase.Config) error) *DefaultValidator {
	return &DefaultValidator{rules: rules}
}

func (d *DefaultValidator) Validate(config *zbase.Config) error {
	for _, ruleFunc := range d.rules {
		if err := ruleFunc(config); err != nil {
			return err
		}
	}
	return nil
}
