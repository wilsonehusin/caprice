package scribe

type TestFataler interface {
	Fatal(args ...interface{})
}

type ScribeT struct {
	Scribe
	testingT TestFataler
}

func NewT(t TestFataler, bucket string) (*ScribeT, error) {
	realScribe, err := New(bucket)
	if err != nil {
		return nil, err
	}
	return &ScribeT{testingT: t, Scribe: *realScribe}, nil
}

func (s *ScribeT) RunT(name string, stagedFunc func() error) {
	s.sendEvent(eventTypeStart, name)
	if err := stagedFunc(); err != nil {
		s.sendEvent(eventTypeFail, name)
		s.testingT.Fatal(err)
	}
	s.sendEvent(eventTypeSuccess, name)
}

func (s *ScribeT) NewStageT(name string) func(error) {
	s.sendEvent(eventTypeStart, name)
	return func(err error) {
		if err != nil {
			s.sendEvent(eventTypeFail, name)
			s.testingT.Fatal(err)
		}
		s.sendEvent(eventTypeSuccess, name)
	}
}
