package client

import "github.com/spf13/cobra"

var dataCmd = &cobra.Command{
	Use:     "data",
	Aliases: []string{"data"},
	Short:   "manage data",
	Long:    `Upload, download, delete and change access on data.`,
}

func init() {
	RootCmd.AddCommand(dataCmd)
}
