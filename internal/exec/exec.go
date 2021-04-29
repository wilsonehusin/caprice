package exec

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/wilsonehusin/caprice/scribe"
)

type ExecOptions struct {
	Destination string
	Source      string
	Bucket      string
	LogDir      string
	Timeout     string
}

func Run(opts *ExecOptions, args []string) error {
	scribe.Destination = opts.Destination
	scribe.SetSource(opts.Source)

	s, err := scribe.New(opts.Bucket)
	if err != nil {
		return err
	}

	// This looks hacky for a reason: the value of scribeErr has to be lazy-evaluated
	var scribeErr error
	defer func() { s.Done(scribeErr) }()

	s.Tags["ExecOptions"] = opts
	s.Tags["args"] = args

	setupEnvDone := s.NewStage("setup environment")

	if len(args) == 0 {
		return fmt.Errorf("no command was specified")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

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

	done := make(chan bool)
	go func() {
		scribeErr = s.RunErr("executing command", cmd.Run)
		done <- true
	}()

	select {
	case <-done:
	case <-ctx.Done():
		log.Info().Msg("allowing graceful shutdown with 3s timeout")
		select {
		case <-done:
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
