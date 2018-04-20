package client

import (
	"github.com/spf13/cobra"
	"log"
	"fmt"
	"os"
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"github.com/racin/DATMAS_2018_Implementation/types"
	tmtypes "github.com/tendermint/tendermint/types"
	"io"
	"encoding/json"
	"bytes"
	"github.com/racin/DATMAS_2018_Implementation/client"
	"time"
)

const newBlockTimeout = 30 * time.Second
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

		// Phase 1. Upload data to Consensus and Storage nodes.
		stranc := getSignedTransaction(types.TransactionType_UploadData, fileHash)
		byteArr, _ := json.Marshal(stranc)
		values := map[string]io.Reader{
			"file":    file,
			"transaction": bytes.NewReader(byteArr),
		}
		res := getAPI().UploadData(&values)
		if res.Codetype != types.CodeType_OK {
			log.Fatal("Error with upload. ", res.Message)
		}

		// Phase 1b. Generate a sample for the file
		stat, _ := file.Stat()
		fileBytes := make([]byte, stat.Size())
		file.Read(fileBytes)
		storageSample := crypto.GenerateStorageSample(&fileBytes)

		// Phase 2. Verify the uploaded data is commited to the ledger
		newBlockCh := make(chan interface{}, 1)
		if err := subToNewBlock(newBlockCh); err != nil {
			log.Fatal("Could not subscribe to new block events. Error: ", err.Error())
		}

		result, err := getAPI().VerifyUpload(stranc)
		if result != types.CodeType_OK {
			log.Fatal("Error verifying upload. ", res.Message)

		}

		castedTx := tmtypes.Tx(byteArr)
		fileName := args[1];
		var fileDescription string;
		if len(args) > 1 {
			fileDescription = args[2]
		}
		// Start timeout to wait for the transaction be put on the ledger.
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
					client.WriteMetadata(fileHash, &client.MetadataEntry{Name:fileName, Description:fileDescription, StorageSample: *storageSample})
				}
			case <-time.After(newBlockTimeout):
				fmt.Println("File was uploaded, but could not verify the ledger within the timeout. " +
					"Try running a status query with CID: " + fileHash)
		}
	},
}

func init() {
	dataCmd.AddCommand(uploadCmd)
}