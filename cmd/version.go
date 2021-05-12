package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/wilsonehusin/caprice/internal/buildinfo"
)

var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"buildinfo"},
	Short:   "Build versions",
	Long:    `Versions related to this build`,
	Run: func(*cobra.Command, []string) {
		versions := *buildinfo.All()
		fmt.Printf("%v: %v\n", "Version", versions["Version"])
		for k, v := range versions {
			if k == "Version" {
				continue
			}
			fmt.Printf("%v: %v\n", k, v)
		}
	},
}
