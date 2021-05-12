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
	EventTypePulse  = "run.caprice.pulse"
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
	client    Sender
	ctx       context.Context
	stopPulse context.CancelFunc
	Bucket    string                 `json:"bucket"`
	Errors    map[string]string      `json:"errors"`
	Tags      map[string]string      `json:"tags"`
	Metadata  map[string]interface{} `json:"metadata"`
}

func New(bucket string) (*Scribe, error) {
	return NewWithContext(bucket, context.Background())
}

func NewWithContext(bucket string, ctx context.Context) (*Scribe, error) {
	client, err := NewSender()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(ctx)
	s := &Scribe{
		client:    client,
		ctx:       ctx,
		stopPulse: cancel,
		Bucket:    bucket,
		Errors:    map[string]string{},
		Tags:      map[string]string{},
		Metadata:  map[string]interface{}{},
	}

	go s.pulse(ctx, "root")

	s.sendEvent(EventTypeStart, "root")
	return s, nil

}

func (s *Scribe) pulse(ctx context.Context, name string) {
	ticker := time.NewTicker(1 * time.Minute)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.sendEvent(EventTypePulse, name)
		}
	}
}

func (s *Scribe) Done(err error) error {
	defer s.sendEvent(EventTypeFinish, "root")
	defer s.stopPulse()

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

	ctx, stopPulse := context.WithCancel(s.ctx)

	go s.pulse(ctx, name)
	defer stopPulse()

	stagedFunc()
}

func (s *Scribe) RunErr(name string, stagedFunc func() error) error {
	s.sendEvent(EventTypeStart, name)
	defer s.sendEvent(EventTypeFinish, name)

	ctx, stopPulse := context.WithCancel(s.ctx)

	go s.pulse(ctx, name)
	defer stopPulse()

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
	ctx, stopPulse := context.WithCancel(s.ctx)

	go s.pulse(ctx, name)

	return func() {
		stopPulse()
		s.sendEvent(EventTypeFinish, name)
	}
}

type ScribeEventData struct {
	Scribe
	Name      string `json:"name"`
	RuntimeID string `json:"runtime_id"`
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
