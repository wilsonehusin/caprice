package exec

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"go.husin.dev/caprice/internal/buildinfo"
	"go.husin.dev/caprice/scribe"
)

type ExecOptions struct {
	Destination string `json:"destination,omitempty"`
	Source      string `json:"source,omitempty"`
	Bucket      string `json:"bucket,omitempty"`
	LogDir      string `json:"logDir,omitempty"`
	Timeout     string `json:"timeout,omitempty"`
}

func Run(opts *ExecOptions, args []string) error {
	if opts.Destination == "" {
		scribe.Destination = buildinfo.Server
	} else {
		scribe.Destination = opts.Destination
	}
	if opts.Source == "" {
		hostname, err := os.Hostname()
		if err != nil {
			scribe.SetSource("exec.caprice")
		}
		scribe.SetSource(hostname)
	} else {
		scribe.SetSource(opts.Source)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	s, err := scribe.NewWithContext(opts.Bucket, ctx)
	if err != nil {
		return err
	}

	// This looks hacky for a reason: the value of scribeErr has to be lazy-evaluated
	var scribeErr error
	defer func() { s.Done(scribeErr) }()

	s.Metadata["options"] = opts
	s.Tags["args"] = strings.Join(args, " ")

	setupEnvDone := s.NewStage("setup environment")

	if len(args) == 0 {
		return fmt.Errorf("no command was specified")
	}

	streams, err := NewExecStreams(opts.LogDir)
	if err != nil {
		scribeErr = err
		return err
	}
	defer streams.Close()

	if opts.Timeout != "" {
		dur, err := time.ParseDuration(opts.Timeout)
		if err != nil {
			scribeErr = err
			return err
		}
		ctx, _ = context.WithTimeout(ctx, dur)
	}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	streams.AttachCmd(cmd)

	setupEnvDone()

	cmdDone := make(chan bool)
	go func() {
		scribeErr = s.RunErr("executing command", cmd.Run)
		cmdDone <- true
	}()

	select {
	case <-cmdDone:
	case <-ctx.Done():
		log.Info().Msg("allowing graceful shutdown with 3s timeout")
		select {
		case <-cmdDone:
		case <-time.After(3 * time.Second):
			if scribeErr == nil {
				scribeErr = fmt.Errorf("timeout waiting for command to shutdown")
			} else {
				scribeErr = fmt.Errorf("timeout waiting for command to shutdown: %w", scribeErr)
			}
		}
	}
	return scribeErr
}
