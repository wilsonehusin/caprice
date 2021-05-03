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
	EventTypeStart  = "run.caprice.start"
	EventTypeFinish = "run.caprice.finish"
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
	client   Sender
	Bucket   string
	Errors   map[string]string
	Tags     map[string]string
	Metadata map[string]interface{}
}

func New(bucket string) (*Scribe, error) {
	client, err := NewSender()
	if err != nil {
		return nil, err
	}
	s := &Scribe{
		client:   client,
		Bucket:   bucket,
		Errors:   map[string]string{},
		Tags:     map[string]string{},
		Metadata: map[string]interface{}{},
	}
	s.sendEvent(EventTypeStart, "root")
	return s, nil
}

func (s *Scribe) Done(err error) error {
	defer s.sendEvent(EventTypeFinish, "root")

	if err != nil {
		s.Errors["root"] = err.Error()
		s.Tags["result"] = "error"
	} else {
		s.Tags["result"] = "success"
	}
	return err
}

func (s *Scribe) Run(name string, stagedFunc func()) {
	s.sendEvent(EventTypeStart, name)
	defer s.sendEvent(EventTypeFinish, name)

	stagedFunc()
}

func (s *Scribe) RunErr(name string, stagedFunc func() error) error {
	s.sendEvent(EventTypeStart, name)
	defer s.sendEvent(EventTypeFinish, name)

	err := stagedFunc()
	if err != nil {
		s.Errors[name] = err.Error()
		s.Tags["result"] = "error"
	} else {
		s.Tags["result"] = "success"
	}
	return err
}

func (s *Scribe) NewStage(name string) func() {
	s.sendEvent(EventTypeStart, name)
	return func() {
		s.sendEvent(EventTypeFinish, name)
	}
}

type ScribeEventData struct {
	Scribe
	Name      string
	RuntimeID string
}

func (s *ScribeEventData) CanonicalName() string {
	return strings.Join([]string{s.RuntimeID, s.Bucket, s.Name}, "-")
}

func (s *Scribe) sendEvent(EventType string, eventName string) {
	e := baseEvent.Clone()
	e.SetType(EventType)
	e.SetTime(time.Now())
	// TODO: handle error?
	_ = e.SetData("application/json", ScribeEventData{
		Scribe:    *s,
		Name:      eventName,
		RuntimeID: runtimeID,
	})
	e.SetID(uuid.New().String())
	// TODO: send in non-blocking goroutine call?
	_ = s.client.Send(context.Background(), e)
}
