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
	if err := stagedFunc(); err != nil {
		s.testingT.Fatal(err)
	}
}

func (s *ScribeT) NewStageT(string) func(error) {
	return func(err error) {
		if err != nil {
			s.testingT.Fatal(err)
		}
	}
}
