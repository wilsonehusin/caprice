package scribe

import (
	"context"
	"encoding/json"
	"io"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type Sender interface {
	Send(ctx context.Context, event cloudevents.Event) cloudevents.Result
}

type StreamSender struct {
	Dest io.Writer
}

func (s *StreamSender) Send(ctx context.Context, event cloudevents.Event) cloudevents.Result {
	data, err := json.Marshal(event)
	if err != nil {
		return cloudevents.Result(err)
	}
	data = append(data, []byte("\n")...)
	_, err = s.Dest.Write(data)
	if err != nil {
		return cloudevents.Result(err)
	}

	return nil
}
