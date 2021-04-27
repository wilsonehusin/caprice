package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/rs/zerolog/log"
)

type Server struct {
	opts ServerOptions
}

type ServerOptions struct {
	Port int `default:"8080" required:"true"`
}

func Run(opts ServerOptions) error {
	client, err := cloudevents.NewClientHTTP(
		cloudevents.WithPort(opts.Port),
	)
	if err != nil {
		return fmt.Errorf("starting server: %w", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	receiverStop := make(chan bool)

	go func() {
		log.Info().Int("port", opts.Port).Msg("starting receiver")
		if err := client.StartReceiver(ctx, processEvent); err != nil {
			log.Err(err).Err(ctx.Err()).Msg("shutdown listener")
			receiverStop <- true
		}
	}()

	select {
	case <-ctx.Done():
		log.Err(ctx.Err()).Msg("context cancelled")
	case <-receiverStop:
		log.Info().Msg("receiver shut down")
	}
	cancel()

	return nil
}

func processEvent(e cloudevents.Event) {
	eventData, err := json.Marshal(e)
	if err != nil {
		log.Error().Err(err).Msg("unable to convert CloudEvent as JSON")
		return
	}
	log.Info().RawJSON("cloudevent", eventData).Send()
}
