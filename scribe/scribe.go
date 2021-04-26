package scribe

import (
	"context"
	"encoding/base32"
	"fmt"
	"io"
	"math"
	"math/rand"
	"strings"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
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

func genRandomB32(length int) string {
	values := make([]byte, length)
	for i := range values {
		values[i] = uint8(rand.Intn(math.MaxUint8))
	}
	return base32.StdEncoding.EncodeToString(values)
}

func init() {
	rand.Seed(time.Now().UnixNano())
	runtimeId = genRandomB32(10)

	baseEvent.SetSpecVersion("1.0")
	baseEvent.SetSource("/caprice/scribe")
}

func SetSource(str string) {
	src := strings.TrimSuffix(str, "/")
	baseEvent.SetSource(src)
}

type Scribe struct {
	client Sender
	bucket string
	Tags   map[string]string
}

func New(bucket string) (*Scribe, error) {
	client, err := NewSender()
	if err != nil {
		return nil, err
	}
	s := &Scribe{
		bucket: bucket,
		client: client,
		Tags:   map[string]string{},
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
	sourcePrefix := []string{e.Source(), s.bucket}
	completeSource := strings.Join(append(sourcePrefix, sourceSuffix...), "/")
	e.SetSource(completeSource)
	e.SetType(eventType)
	e.SetTime(time.Now())
	e.SetData("application/json", s.Tags)
	e.SetID(fmt.Sprintf("r-%s-e-%s", runtimeId, genRandomB32(10)))
	// TODO: send in non-blocking goroutine call?
	_ = s.client.Send(context.Background(), e)
}
