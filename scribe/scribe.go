package scribe

import (
	"context"
	"os"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

var (
	baseEvent = cloudevents.NewEvent()

	tags = map[string]string{}
)

const (
	eventTypeStart   = "run.caprice.start"
	eventTypeFinish  = "run.caprice.finish"
	eventTypeSuccess = "run.caprice.success"
	eventTypeFail    = "run.caprice.fail"
)

func init() {
	baseEvent.SetSpecVersion("1.0")
	baseEvent.SetSource("/caprice/scribe")
	baseEvent.SetDataContentType("application/json")
}

func SetSource(str string) {
	src := strings.TrimSuffix(str, "/")
	baseEvent.SetSource(src)
}

type Scribe struct {
	client    Sender
	bucket    string
	ExtraTags map[string]string
}

func New(bucket string) (*Scribe, error) {
	client := &StreamSender{Dest: os.Stdout}
	s := &Scribe{
		bucket: bucket,
		client: client,
	}
	s.sendEvent(eventTypeStart)
	return s, nil
}

func (s *Scribe) Run(name string, stagedFunc func()) {
	s.sendEvent(eventTypeStart, name)
	stagedFunc()
	s.sendEvent(eventTypeFinish, name)
}

func (s *Scribe) RunErr(name string, stagedFunc func() error) error {
	s.sendEvent(eventTypeStart, name)
	eventStatus := eventTypeSuccess
	err := stagedFunc()
	if err != nil {
		eventStatus = eventTypeFail
	}
	s.sendEvent(eventStatus, name)
	return err
}

func (s *Scribe) Done(err error) {
	eventStatus := eventTypeSuccess
	if err != nil {
		eventStatus = eventTypeFail
	}
	s.sendEvent(eventStatus)
}

func (s *Scribe) NewStage(name string) func() {
	s.sendEvent(eventTypeStart, name)
	return func() {
		s.sendEvent(eventTypeFinish, name)
	}
}

func (s *Scribe) sendEvent(eventType string, sourceSuffix ...string) {
	e := baseEvent.Clone()
	sourcePrefix := []string{e.Source()}
	completeSource := strings.Join(append(sourcePrefix, sourceSuffix...), "/")
	e.SetSource(completeSource)
	e.SetType(eventType)
	e.SetTime(time.Now())
	// TODO: send in non-blocking goroutine call?
	_ = s.client.Send(context.Background(), e)
}
