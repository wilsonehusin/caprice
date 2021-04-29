package scribe

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

type Sender interface {
	Send(ctx context.Context, event cloudevents.Event) cloudevents.Result
}

type StreamSender struct {
	Dest io.Writer
}

func NewSender() (Sender, error) {
	var s Sender
	switch {
	case DestinationWriter != nil:
		s = &StreamSender{Dest: DestinationWriter}
	case strings.HasPrefix(Destination, "http://"), strings.HasPrefix(Destination, "https://"):
		client, err := cloudevents.NewClientHTTP(
			cloudevents.WithTarget(Destination),
		)
		if err != nil {
			return nil, err
		}
		s = client
	case strings.HasPrefix(Destination, "file://"):
		// TODO: open file, pass the file to StreamSender
		fallthrough
	case Destination == "stderr":
		s = &StreamSender{Dest: os.Stderr}
	case Destination == "stdout":
		s = &StreamSender{Dest: os.Stderr}
	default:
		s = &StreamSender{Dest: ioutil.Discard}
	}
	return s, nil
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
