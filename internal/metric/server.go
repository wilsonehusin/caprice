package metric

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/rs/zerolog/log"
)

func StartMetricsServer(ctx context.Context, cancel func(), port int, pushGateway string) {
	pusher = push.New(pushGateway, "caprice_scribe")

	server := &http.Server{Addr: fmt.Sprintf(":%d", port)}
	http.Handle("/metrics", promhttp.Handler())

	serverHasShutdown := make(chan bool)
	go func(server *http.Server) {
		log.Info().Int("port", port).Msg("starting metrics endpoint")
		err := server.ListenAndServe()
		log.Err(err).Int("port", port).Msg("metrics endpoint shut down")
		serverHasShutdown <- true
	}(server)

	go func() {
		<-serverHasShutdown
		cancel()
	}()

	go trackScribeTimeout(ctx)

	<-ctx.Done()
	log.Info().Msg("shutting down metrics endpoint")
	ctxServer, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := server.Shutdown(ctxServer)
	log.Err(err).Msg("metrics endpoint shut down")
}
