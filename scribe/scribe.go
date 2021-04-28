package scribe

import (
	"context"
	"io"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
)

var (
	baseEvent = cloudevents.NewEvent()

	DestinationWriter io.Writer
	Destination       string
	runtimeId         string
)

const (
	eventTypeStart   = "run.caprice.start"
	eventTypeFinish  = "run.caprice.finish"
	eventTypeSuccess = "run.caprice.success"
	eventTypeFail    = "run.caprice.fail"
)

func init() {
	runtimeId = uuid.New().String()

	baseEvent.SetSpecVersion("1.0")
	baseEvent.SetSource("")
}

func SetSource(str string) {
	src := strings.TrimSuffix(str, "/")
	baseEvent.SetSource(src)
}

type Scribe struct {
	client Sender
	bucket string
	Tags   map[string]interface{}
}

func New(bucket string) (*Scribe, error) {
	client, err := NewSender()
	if err != nil {
		return nil, err
	}
	s := &Scribe{
		bucket: bucket,
		client: client,
		Tags:   map[string]interface{}{},
	}
	s.sendEvent(eventTypeStart, "")
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
	s.sendEvent(eventStatus, "")
}

func (s *Scribe) NewStage(name string) func() {
	s.sendEvent(eventTypeStart, name)
	return func() {
		s.sendEvent(eventTypeFinish, name)
	}
}

func (s *Scribe) sendEvent(eventType string, eventName string) {
	e := baseEvent.Clone()
	e.SetType(eventType)
	e.SetTime(time.Now())
	// TODO: handle error?
	_ = e.SetData("application/json", map[string]interface{}{
		"tags":    s.Tags,
		"runtime": runtimeId,
		"bucket":  s.bucket,
		"name":    eventName,
	})
	e.SetID(uuid.New().String())
	// TODO: send in non-blocking goroutine call?
	_ = s.client.Send(context.Background(), e)
}
