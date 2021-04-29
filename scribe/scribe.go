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
	runtimeID         string
)

const (
	EventTypeStart   = "run.caprice.start"
	EventTypeFinish  = "run.caprice.finish"
	EventTypeSuccess = "run.caprice.success"
	EventTypeFail    = "run.caprice.fail"
)

func init() {
	runtimeID = uuid.New().String()

	baseEvent.SetSpecVersion("1.0")
	baseEvent.SetSource("")
}

func RuntimeID() string {
	return runtimeID
}

func SetSource(str string) {
	src := strings.TrimSuffix(str, "/")
	baseEvent.SetSource(src)
}

type Scribe struct {
	bucket string
	client Sender
	errors map[string]interface{}
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
		errors: map[string]interface{}{},
		Tags:   map[string]interface{}{},
	}
	s.sendEvent(EventTypeStart, "root")
	return s, nil
}

func (s *Scribe) Run(name string, stagedFunc func()) {
	s.sendEvent(EventTypeStart, name)
	stagedFunc()
	s.sendEvent(EventTypeFinish, name)
}

func (s *Scribe) RunErr(name string, stagedFunc func() error) error {
	s.sendEvent(EventTypeStart, name)
	eventStatus := EventTypeSuccess
	err := stagedFunc()
	if err != nil {
		eventStatus = EventTypeFail
		s.errors[name] = err.Error()
	}
	s.sendEvent(eventStatus, name)
	return err
}

func (s *Scribe) Done(err error) error {
	eventStatus := EventTypeSuccess
	if err != nil {
		eventStatus = EventTypeFail
		s.errors["root"] = err.Error()
	}
	s.sendEvent(eventStatus, "root")
	return err
}

func (s *Scribe) NewStage(name string) func() {
	s.sendEvent(EventTypeStart, name)
	return func() {
		s.sendEvent(EventTypeFinish, name)
	}
}

func (s *Scribe) sendEvent(EventType string, eventName string) {
	e := baseEvent.Clone()
	e.SetType(EventType)
	e.SetTime(time.Now())
	// TODO: handle error?
	_ = e.SetData("application/json", map[string]interface{}{
		"tags":    s.Tags,
		"runtime": runtimeID,
		"bucket":  s.bucket,
		"name":    eventName,
		"errors":  s.errors,
	})
	e.SetID(uuid.New().String())
	// TODO: send in non-blocking goroutine call?
	_ = s.client.Send(context.Background(), e)
}
