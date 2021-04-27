package scribe

import (
	"fmt"
	"testing"
)

type fakeTestingT struct {
	AttachedErrs []interface{}
	Called       bool
}

func (f *fakeTestingT) Fatal(errs ...interface{}) {
	f.Called = true
	f.AttachedErrs = errs
}

func TestTriggerFatal(t *testing.T) {
	fakeT := &fakeTestingT{}

	s, _ := NewT(fakeT, "TestMyAppTriggerFatal")

	err := fmt.Errorf("expected failure")
	s.RunT("thisShouldFail", func() error {
		return err
	})
	s.Done(nil)

	if !fakeT.Called {
		t.Fatalf("expected Fatal() to be called, but was not")
	}

	if len(fakeT.AttachedErrs) == 0 {
		t.Fatalf("expected to receive 1 error, got none")
	}

	if fakeT.AttachedErrs[0].(error) != err {
		t.Fatalf("Fatal() was called with a different error")
	}
}

func TestNoTriggerFatal(t *testing.T) {
	fakeT := &fakeTestingT{}

	s, _ := NewT(fakeT, "TestMyAppNoTriggerFatal")

	s.RunT("thisShouldNotFail", func() error {
		return nil
	})
	s.Done(nil)

	if fakeT.Called {
		t.Fatalf("unexpected call to Fatal()")
	}
}
