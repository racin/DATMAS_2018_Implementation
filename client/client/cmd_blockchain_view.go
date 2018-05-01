package client

import (
	"github.com/spf13/cobra"
	"log"
	"strconv"
	"fmt"
	"bytes"
	"encoding/json"
)

var blockchainViewCmd = &cobra.Command{
	Use:     "view [height]",
	Short:   "print block contents",
	Long:    `Prints the transaction data contained in a block.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			log.Fatal("Not enough arguments.")
		}
		height, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatal("Invalid input.")
		}
		height64 := int64(height)
		result, err := TheClient.TMClient.Block(&height64);
		if err != nil {
			log.Fatal("Error: " + err.Error())
		}
		if err := result.Block.ValidateBasic(); err != nil {
			log.Fatal("Could not validate block. Error: ", err.Error())
		}
		for i := int64(0); i < result.Block.NumTxs; i++ {
			txData := []byte(result.Block.Txs[i])
			var prettyJSON bytes.Buffer
			err = json.Indent(&prettyJSON, txData, "", "\t")
			if err != nil {
				continue // Problems with one transaction. Continue
			}
			fmt.Printf("-------------\nTransaction %v:\n%v\n", i, string(prettyJSON.Bytes()))
		}
	},
}

func init() {
	blockchainCmd.AddCommand(blockchainViewCmd)
}
