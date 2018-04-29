package client

import (
	"github.com/spf13/cobra"
	"log"
)

var dataCmd = &cobra.Command{
	Use:     "challenge [CID] [challenge]",
	Aliases: []string{"challenge"},
	Short:   "Challenge storage nodes",
	Long:    `Challenge storage nodes to prove that they still possess all the data for a CID.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			log.Fatal("Not enough arguments.")
		}
	},
}

func init() {
	RootCmd.AddCommand(dataCmd)
}
