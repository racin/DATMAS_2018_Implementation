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
	"time"
	"io/ioutil"
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


		fmt.Printf("TheClient: %+v\n", TheClient)
		// File and Name is required parameters.
		filePath := args[0];

		// TODO: Figure out if there is a way to only open the file once.
		file, err := os.Open(filePath)
		defer file.Close()
		if err != nil {
			log.Fatal("Could not open file. Error: ", err.Error())
		}
		file2, err := os.Open(filePath)
		defer file2.Close()
		if err != nil {
			log.Fatal("Could not open file. Error: ", err.Error())
		}
		fileBytes, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Fatal("Could not get file bytes. Error: " + err.Error())
		}
		fileHash, err := crypto.IPFSHashData(fileBytes)
		if err != nil {
			log.Fatal("Could not hash file. Error: ", err.Error())
		}

		// Phase 1. Upload data to Consensus and Storage nodes.
		stranc := getSignedTransaction(types.TransactionType_UploadData, &types.RequestUpload{Cid:fileHash, IpfsNode:TheClient.IPFSIdent})
		fmt.Printf("Tranc: %+v\n", stranc.Base.(*types.Transaction))
		byteArr, _ := json.Marshal(stranc)
		valuesTM := map[string]io.Reader{
			"file":    file,
			"transaction": bytes.NewReader(byteArr),
		}
		valuesIPFS := map[string]io.Reader{
			"file":    file2,
			"transaction": bytes.NewReader(byteArr),
		}


		if res := TheClient.UploadDataToTM(&valuesTM); res.Codetype != types.CodeType_OK {
			log.Fatal("Error with TM upload. ", res.Message)
		}
		if res := TheClient.UploadDataToIPFS(&valuesIPFS); res.Codetype != types.CodeType_OK {
			log.Fatal("Error with IPFS upload. ", string(res.Message))
		}
		// Phase 1b. Generate a sample for the file
		storageSample := crypto.GenerateStorageSample(&fileBytes)

		// Phase 2. Verify the uploaded data is commited to the ledger
		newBlockCh := make(chan interface{}, 1)
		if err := subToNewBlock(newBlockCh); err != nil {
			log.Fatal("Could not subscribe to new block events. Error: ", err.Error())
		}


		if _, err := TheClient.VerifyUpload(stranc); err != nil {
			log.Fatal("Error verifying upload. Error: " + err.Error())
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
				evt := b.(tmtypes.EventDataNewBlock)
				// Validate
				if err := evt.Block.ValidateBasic(); err != nil {
					// System is broken. Notify administrators
					log.Fatal("Could not validate latest block. Error: ", err.Error())
				}
				if evt.Block.Txs.Index(castedTx) > -1 {
					// Transaction is put in the latest block.
					fmt.Println("File successfully uploaded. CID: ", fileHash)
					WriteMetadata(fileHash, &MetadataEntry{Name:fileName, Description:fileDescription, StorageSample: *storageSample})
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