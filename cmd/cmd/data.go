package cmd

import "github.com/spf13/cobra"

var dataCmd = &cobra.Command{
	Use:     "data",
	Aliases: []string{"data"},
	Short:   "manage data",
	Long:    `Upload, download and remove data.`,
}

func init() {
	RootCmd.AddCommand(dataCmd)
}
