package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "worker",
		Short: "Worker",
		Long:  "Worker",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO
		},
	})
}
