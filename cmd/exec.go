package cmd

import (
	"bytes"

	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"go.husin.dev/caprice/internal/exec"
)

var execOpts = &exec.ExecOptions{}

var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Executes binary and send CloudEvents on start and finish",
	Long: `Executes binary specified in arguments while reporting the execution time
as CloudEvents. This practically wraps the execution in scribe.RunErr() while
still forwarding Stdout and Stderr as if the program is exec natively.`,
	PreRun: loadexecOpts,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		return exec.Run(execOpts, args)
	},
}

func loadexecOpts(cmd *cobra.Command, args []string) {
	if err := envconfig.Process(rootCmdName, execOpts); err != nil {
		log.Fatal().Err(err).Msg("failed to process exec options")
	}
}

func init() {
	var optionsUsage bytes.Buffer
	if err := envconfig.Usagef(rootCmdName, execOpts, &optionsUsage, optionsUsageTemplate); err != nil {
		log.Fatal().Err(err).Msg("failed to retrieve exec options usage")
	}
	execCmd.SetUsageTemplate(execCmd.UsageTemplate() + optionsUsage.String() + "\n")
}
