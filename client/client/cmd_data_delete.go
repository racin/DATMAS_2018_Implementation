package client

import (
	"github.com/spf13/cobra"
	"log"
	"fmt"
	"github.com/racin/DATMAS_2018_Implementation/types"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	tmtypes "github.com/tendermint/tendermint/types"
	"encoding/json"
	"time"
)

// getAccountCmd represents the getAccount command
var dataRemoveCmd = &cobra.Command{
	Use:   "delete [CID]",
	Short: "Delete data",
	Long:  `Delete data from the storage network.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			log.Fatal("Not enough arguments.")
		}

		cid := args[0];
		stx := TheClient.GetSignedTransaction(types.TransactionType_RemoveData, cid)

		// Start listening for new block
		newBlockCh := make(chan interface{}, 1)
		if err := TheClient.SubToNewBlock(newBlockCh); err != nil {
			log.Fatal("Could not subscribe to new block events. Error: ", err.Error())
		}

		stxByteArr, err := json.Marshal(stx)
		if err != nil {
			log.Fatal("Error marshalling: Error: " + err.Error())
		}

		castedTx := tmtypes.Tx(stxByteArr)
		if _, err := types.CheckBroadcastResult(TheClient.TMClient.BroadcastTxSync(castedTx)); err != nil {
			log.Fatal("Error broadcasting request. Error: " + err.Error())
		}

		// Start timeout to wait for the transaction be put on the ledger.
		select {
		case b := <-newBlockCh:
			evt := b.(tmtypes.EventDataNewBlock)
			if err := evt.Block.ValidateBasic(); err != nil {
				// System is broken. Notify administrators
				log.Fatal("Could not validate latest block. Error: ", err.Error())
			}
			if evt.Block.Txs.Index(castedTx) > -1 {
				// Transaction is put in the latest block.
				if err := types.DeleteMetadata(cid); err == nil {
					fmt.Println("File was successfully deleted. Metadata deleted.")
				} else {
					fmt.Println("File was successfully deleted. Could not delete metadata. Error: " + err.Error())
				}
			}
		case <-time.After(time.Duration(conf.ClientConfig().NewBlockTimeout) * time.Second):
			fmt.Println("Could not verify the ledger within the timeout.")
		}
	},
}

func init() {
	dataCmd.AddCommand(dataRemoveCmd)
}