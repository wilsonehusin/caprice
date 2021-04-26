package cmd

import (
	"bytes"

	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/wilsonehusin/caprice/internal/server"
)

var serverOpts server.ServerOptions

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "HTTP server listening and aggregating CloudEvents",
	Long: `HTTP server listening and aggregating CloudEvents produced by scribe.
By default, data is stored in-memory, which should not be used in production.`,
	PreRun: loadServerOpts,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		return server.Run(serverOpts)
	},
}

func loadServerOpts(cmd *cobra.Command, args []string) {
	if err := envconfig.Process(rootCmdName, &serverOpts); err != nil {
		log.Fatal().Err(err).Msg("failed to process server options")
	}
}

func init() {
	var optionsUsage bytes.Buffer
	if err := envconfig.Usagef(rootCmdName+"_server", &serverOpts, &optionsUsage, optionsUsageTemplate); err != nil {
		log.Fatal().Err(err).Msg("failed to retrieve server options usage")
	}
	serverCmd.SetUsageTemplate(serverCmd.UsageTemplate() + optionsUsage.String() + "\n")
}
