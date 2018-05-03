package client

import (
	"github.com/spf13/cobra"
	"log"
	"fmt"
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"github.com/racin/DATMAS_2018_Implementation/types"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	tmtypes "github.com/tendermint/tendermint/types"
	"encoding/json"
	"time"
	"strings"
)

// getAccountCmd represents the getAccount command
var dataAccessCmd = &cobra.Command{
	Use:   "access [READERS]",
	Short: "Readers is a comma separated list",
	Long:  `Enable download of data for the identities given as a comma separated list.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			log.Fatal("Not enough arguments.")
		}

		readers := strings.Split(",", args[0])
		tx := types.NewTx(readers, TheClient.fingerprint, types.TransactionType_ChangeContentAccess)
		stx, err := crypto.SignStruct(tx, TheClient.privKey)
		if err != nil {
			log.Fatal(err.Error())
		}

		// Start listening for new block
		// Phase 2. Send metadata to TM
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
			// Validate
			if err := evt.Block.ValidateBasic(); err != nil {
				// System is broken. Notify administrators
				log.Fatal("Could not validate latest block. Error: ", err.Error())
			}
			if evt.Block.Txs.Index(castedTx) > -1 {
				// Transaction is put in the latest block.
				fmt.Println("Readers of data successfully recorded on the ledger.")
			}
		case <-time.After(time.Duration(conf.ClientConfig().NewBlockTimeout) * time.Second):
			fmt.Println("Could not verify the ledger within the timeout.")
		}
	},
}

func init() {
	dataCmd.AddCommand(dataAccessCmd)
}