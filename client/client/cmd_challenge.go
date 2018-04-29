package client

import "github.com/spf13/cobra"

var challegeCmd = &cobra.Command{
	Use:     "challenge",
	Aliases: []string{"data"},
	Short:   "manage data",
	Long:    `Upload, download and remove data.`,
}

func init() {
	RootCmd.AddCommand(challegeCmd)
}
