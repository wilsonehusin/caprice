package scribe

import (
	"fmt"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

var (
	Source      string
	Destination string
)

var (
	eventTypePrefix = []string{
		"run",
		"caprice",
	}
)

type Scribe struct {
	bucket string
	client cloudevents.Client
	config ScribeConfig
}

type ScribeConfig struct {
	Source      string
	Destination string
	Type        string
}

func New(bucket string) (*Scribe, error) {
	client, err := cloudevents.NewClientHTTP()
	if err != nil {
		return nil, err
	}
	return &Scribe{bucket: bucket, client: client}, nil
}

func (s *Scribe) Run(name string, stagedFunc func()) {
	stagedFunc()
}

func (s *Scribe) RunErr(name string, stagedFunc func() error) error {
	return stagedFunc()
}

func (s *Scribe) Done(error) {}

func (s *Scribe) NewStage(string) func() {
	return func() {}
}

func (s *Scribe) newCloudEvent(eventType string) *cloudevents.Event {
	event := cloudevents.NewEvent()
	event.SetSource(fmt.Sprintf("caprice/%s", s.bucket))
	event.SetType(eventType)
	return &event
}

func newEventType(strs ...string) string {
	return strings.Join(append(eventTypePrefix, strs...), ".")
}
