package cmd

import (
	"fmt"

	"github.com/MichielBijland/uncomplicated-registry/internal/version"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the version of the Uncomplicated Registry",
	Run: func(cmd *cobra.Command, args []string) {
		md := version.GetVersionMetadata()
		fmt.Println(md.String())
	},
}
