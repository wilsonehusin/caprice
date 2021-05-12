package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/rs/zerolog/log"

	"github.com/wilsonehusin/caprice/internal/metric"
	"github.com/wilsonehusin/caprice/scribe"
)

type ServerOptions struct {
	EventsPort  int    `default:"8080" required:"true"`
	MetricsPort int    `default:"9090" required:"true"`
	PushGateway string `default:"localhost:9091" required:"true"`
}

func Run(opts ServerOptions) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	go metric.StartMetricsServer(ctx, cancel, opts.MetricsPort, opts.PushGateway)

	client, err := cloudevents.NewClientHTTP(
		cloudevents.WithPort(opts.EventsPort),
	)
	if err != nil {
		return fmt.Errorf("starting server: %w", err)
	}

	receiverStop := make(chan bool)
	go func() {
		log.Info().Int("port", opts.EventsPort).Msg("starting events receiver")

		// Blocking call
		err := client.StartReceiver(ctx, processEvent)

		log.Err(err).Msg("stopped events receiver")
		receiverStop <- true
	}()

	<-ctx.Done()
	log.Err(ctx.Err()).Msg("gracefully shutting down (10s timeout)")

	select {
	case <-receiverStop:
		log.Info().Msg("events receiver shut down")
	case <-time.After(10 * time.Second):
		log.Error().Msg("deadline exceeded, forcibly shutting down")
	}

	return nil
}

func processEvent(e cloudevents.Event) {
	var scribeEventData scribe.ScribeEventData
	if err := e.DataAs(&scribeEventData); err != nil {
		log.Error().Err(err).Msg("unrecognized CloudEvent data")
		return
	}

	eventType := e.Type()
	switch eventType {
	case scribe.EventTypeStart:
		metric.IncScribe(scribeEventData.CanonicalName())
	case scribe.EventTypePulse:
		metric.PulseScribe(scribeEventData.CanonicalName())
	case scribe.EventTypeFinish:
		metric.DecScribe(scribeEventData.CanonicalName())
	default:
		log.Error().Str("eventType", eventType).Msg("unrecognized CloudEvent type")
	}

	eventData, err := json.Marshal(e)
	if err != nil {
		log.Error().Err(err).Msg("unable to convert CloudEvent as JSON")
		return
	}
	log.Info().RawJSON("event", eventData).Send()
}
