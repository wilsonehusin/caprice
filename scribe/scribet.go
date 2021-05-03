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
	if err := s.RunErr(name, stagedFunc); err != nil {
		s.testingT.Fatal(err)
	}
}

func (s *ScribeT) NewStageT(name string) func(error) {
	s.sendEvent(EventTypeStart, name)
	return func(err error) {
		defer s.sendEvent(EventTypeFinish, name)
		if err != nil {
			s.Tags["result"] = "fail"
			s.testingT.Fatal(err)
		} else {
			s.Tags["result"] = "pass"
		}
	}
}
