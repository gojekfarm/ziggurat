package ziggurat

import (
	"errors"
	"reflect"
	"testing"
)

func TestJoin(t *testing.T) {
	e1 := make(chan error)
	e2 := make(chan error)
	error1 := errors.New("e1 error")
	error2 := errors.New("e2 error")
	expectedErrorSlice := []error{error1, error2}
	go func() {
		e1 <- error1
		e2 <- error2
	}()
	actualErrorSlice := Join(e1, e2)
	if !reflect.DeepEqual(actualErrorSlice, expectedErrorSlice) {
		t.Errorf("expected error slice %v got %v", expectedErrorSlice, actualErrorSlice)
	}
}
