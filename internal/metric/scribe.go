package metric

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/rs/zerolog/log"

	"go.husin.dev/caprice/internal/scribetag"
)

var (
	pusher *push.Pusher

	scribeTracker = scribetag.NewTracker()

	activeScribeGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "caprice_tracked_scribes",
		Help: "Currently tracked (active) scribes",
	})
)

func trackScribeTimeout(ctx context.Context) {
	timer := time.NewTicker(3 * time.Second)
	for {
		select {
		case <-timer.C:
			logTimeouts()
		case <-ctx.Done():
			return
		}
	}
}

func logTimeouts() {
	var oldest *scribetag.ScribeTag
	for {
		oldest = scribeTracker.Oldest()
		if oldest == nil {
			break
		}
		if oldest.LastPulse.Add(30 * time.Second).After(time.Now()) {
			break
		}
		log.Error().Str("name", oldest.Name).Time("lastPulse", oldest.LastPulse).Msg("timed out")
		DecScribe(oldest.Name)
	}
}

func IncScribe(name string) {
	now := time.Now()
	if err := scribeTracker.Add(name, now); err != nil {
		log.Warn().Err(err).Str("name", name).Msg("not tracking scribe")
	} else {
		activeScribeGauge.Inc()
		log.Info().Str("name", name).Time("lastPulse", now).Msg("increment caprice_tracked_scribes")
	}
}

func PulseScribe(name string) {
	now := time.Now()
	if err := scribeTracker.Pulse(name, now); err != nil {
		log.Warn().Err(err).Str("name", name).Msg("ignoring untracked scribe pulse")
	} else {
		log.Info().Str("name", name).Time("lastPulse", now).Msg("pulse")
	}
}

func DecScribe(name string) {
	if tag := scribeTracker.Take(name); tag != nil {
		activeScribeGauge.Dec()
		log.Info().Str("name", tag.Name).Time("lastPulse", tag.LastPulse).Msg("decrement caprice_tracked_scribes")
	} else {
		log.Warn().Str("name", name).Msg("ignoring untracked scribe deletion")
	}
}
