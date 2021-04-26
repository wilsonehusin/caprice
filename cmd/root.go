package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const (
	rootCmdName = "caprice"

	optionsUsageHeader   = "\nEnvironment Variables (* required):"
	optionsUsageTemplate = `{{range .}}
  {{if usage_required .}}(*) {{else}}    {{end}}{{usage_key .}}={{usage_default .}}{{end}}`
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type rootOptions struct {
	Debug   bool `default:"false"`
	console bool
}

var rootOpts = rootOptions{}

var rootCmd = &cobra.Command{
	Use:              rootCmdName,
	Short:            "Caprice provides visibility to execution time and progress",
	Long:             `Caprice provides visibility to execution time and progress`,
	PersistentPreRun: rootCmdInit,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%+v\n\n", rootOpts)
		cmd.Help()
	},
}

func rootCmdInit(cmd *cobra.Command, args []string) {
	if err := envconfig.Process(rootCmdName, &rootOpts); err != nil {
		log.Fatal().Err(err).Msg("failed to process root options")
	}

	if rootOpts.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = log.With().Caller().Logger()
		zerolog.CallerMarshalFunc = func(file string, line int) string {
			return path.Base(file) + ":" + strconv.Itoa(line)
		}
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	if rootOpts.console {
		log.Logger = log.Logger.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
}

func rootCmdOptionsUsage() string {
	var optionsUsage bytes.Buffer
	if err := envconfig.Usagef(rootCmdName, &rootOpts, &optionsUsage, optionsUsageTemplate); err != nil {
		log.Fatal().Err(err).Msg("failed to retrieve root options usage")
	}
	return optionsUsage.String()
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&rootOpts.Debug, "debug", rootOpts.Debug, "Show debug logs")
	// Intentionally only configurable by flag as it's only useful on active TTY
	rootCmd.PersistentFlags().BoolVar(&rootOpts.console, "console", false, "Use human-readable logs")

	rootCmd.SetUsageTemplate(rootCmd.UsageTemplate() + optionsUsageHeader + rootCmdOptionsUsage() + "\n")

	rootCmd.AddCommand(serverCmd)
}
