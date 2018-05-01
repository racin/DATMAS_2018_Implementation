package client

import "github.com/spf13/cobra"

var blockchainCmd = &cobra.Command{
	Use:     "blockchain",
	Aliases: []string{"blockchain"},
	Short:   "blockchain operations",
	Long:    `Inspect the blockchain.`,
}

func init() {
	RootCmd.AddCommand(blockchainCmd)
}
