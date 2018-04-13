package cmd

import (
	"github.com/spf13/cobra"
	"log"
	"github.com/racin/DATMAS_2018_Implementation/app"
	"fmt"
	"os"
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"github.com/racin/DATMAS_2018_Implementation/types"
	tmtypes "github.com/tendermint/tendermint/types"
	"io"
	"encoding/json"
	"bytes"
)

// getAccountCmd represents the getAccount command
var uploadCmd = &cobra.Command{
	Use:   "upload [file] [name] [description]",
	Short: "upload data",
	Long:  `Upload data.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			log.Fatal("Not enough arguments.")
		}

		// File and Name is required parameters.
		filePath := args[0];
		file, err := os.Open(filePath)
		if err != nil {
			log.Fatal("Could not open file. Error: ", err.Error())
		}

		fileHash, err := crypto.IPFSHashFile(filePath)
		if err != nil {
			log.Fatal("Could not hash file. Error: ", err.Error())
		}

		stranc := getSignedTransaction(app.UploadData, fileHash)
		result, err := getAPI().BeginUploadData(stranc)

		if result != types.CodeType_BCFSBeginUploadOK {
			log.Fatal(err.Error())

		}
		byteArr, _ := json.Marshal(stranc)
		fmt.Println("Data hash registered in application. Uploading data to service")
		values := map[string]io.Reader{
			"file":    file,
			"transaction": bytes.NewReader(byteArr),
		}

		res := getAPI().EndUploadData(&values)

		if res.Codetype != types.CodeType_OK {
			log.Fatal("Error with upload. ", res.Message)
		}

		newBlockCh := make(chan interface{}, 1)
		if err := subToNewBlock(newBlockCh); err != nil {
			log.Fatal("Could not subscribe to new block events. Error: ", err.Error()
		}

		castedTx := tmtypes.Tx(byteArr)
		fileName := args[1];
		var fileDescription string;
		if len(args) > 1 {
			fileDescription = args[2]
		}
		select {
			case b := <-newBlockCh:
				evt := b.(tmtypes.TMEventData).Unwrap().(tmtypes.EventDataNewBlock)
				// Validate
				if err := evt.Block.ValidateBasic(); err != nil {
					// System is broken. Notify administrators
					log.Fatal("Could not validate latest block. Error: ", err.Error())
				}
				if evt.Block.Txs.Index(castedTx) > -1 {
					// Transaction is put in the latest block.

					fmt.Println("File successfully uploaded. CID: ", fileHash)
				}
				nTxs += int(evt.Block.Header.NumTxs)
			case <-ticker.C:
				panic("Timed out waiting to commit blocks with transactions")
		}
		// Start timeout to wait for the transaction be put on the ledger.
	},
}


func init() {
	dataCmd.AddCommand(uploadCmd)
}