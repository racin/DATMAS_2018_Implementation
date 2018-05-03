package client

import (
	"github.com/spf13/cobra"
	"log"
	"fmt"
	"os"
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"github.com/racin/DATMAS_2018_Implementation/types"
	tmtypes "github.com/tendermint/tendermint/types"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	"io"
	"encoding/json"
	"bytes"
	"time"
	"io/ioutil"
)

var uploadCmd = &cobra.Command{
	Use:   "upload [file] [name] [description]",
	Short: "Upload data",
	Long:  `Upload data to the storage network..`,
	Run: func(cmd *cobra.Command, args []string) {
		// File and Name is required parameters.
		if len(args) < 2 {
			log.Fatal("Not enough arguments.")
		}
		filePath := args[0];

		file, err := os.Open(filePath)
		defer file.Close()
		if err != nil {
			log.Fatal("Could not open file. Error: ", err.Error())
		}
		fileStat, err := file.Stat()
		if err != nil {
			log.Fatal("Could not get Stat of file. Error: ", err.Error())
		}

		fileBytes, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Fatal("Could not get file bytes. Error: " + err.Error())
		}
		fileHash, err := crypto.IPFSHashData(fileBytes)
		if err != nil {
			log.Fatal("Could not hash file. Error: ", err.Error())
		}

		// Phase 1. Upload data to one Storage node and verify its signature that it received the data.
		ipfsIdentity, ipfsPubKey := crypto.GetIdentityPublicKey(TheClient.IPFSIdent, TheClient.GetAccessList(),
			conf.ClientConfig().BasePath + conf.ClientConfig().PublicKeys)
		if ipfsIdentity == nil {
			log.Fatal("Could not find IPFS node in the access list.")
		}
		sentReqUpload := &types.RequestUpload{Cid:fileHash, IpfsNode:TheClient.IPFSIdent, Length:fileStat.Size()}
		stranc := TheClient.GetSignedTransaction(types.TransactionType_UploadData, sentReqUpload)
		byteArr, _ := json.Marshal(stranc)
		values := map[string]io.Reader{
			"file":    file,
			"transaction": bytes.NewReader(byteArr),
		}

		ipfsResponse := TheClient.sendMultipartFormDataToIPFS(&values);
		if ipfsResponse.Codetype != types.CodeType_OK {
			log.Fatal("Error with IPFS upload. ", string(ipfsResponse.Message))
		}
		ipfsStx := &crypto.SignedStruct{Base: &types.RequestUpload{}}
		if err := json.Unmarshal(ipfsResponse.Message, ipfsStx); err != nil {
			log.Fatal("Erroneous response from IPFS. Err:", err.Error())
		} else if reqUpload, ok := ipfsStx.Base.(*types.RequestUpload); !ok {
			log.Fatal("Erroneous response from IPFS")
		} else if !reqUpload.CompareTo(sentReqUpload) {
			log.Fatal("Sent upload request is not equal to response.")
		}

		if !ipfsStx.Verify(ipfsPubKey) {
			log.Fatal("Could not verify IPFS signature.")
		}

		// Phase 1b. Generate a sample for the file
		storageSample := crypto.GenerateStorageSample(&fileBytes)
		storageSample.Identity = TheClient.fingerprint

		// Phase 2. Send metadata to TM
		newBlockCh := make(chan interface{}, 1)
		if err := TheClient.SubToNewBlock(newBlockCh); err != nil {
			log.Fatal("Could not subscribe to new block events. Error: ", err.Error())
		}

		strancTM := TheClient.GetSignedTransaction(types.TransactionType_UploadData, ipfsStx)
		byteArrTranc, err := json.Marshal(strancTM)
		if err != nil {
			log.Fatal("Could not generate a byte array of transaction.")
		}

		if _, err := TheClient.VerifyUpload(strancTM); err != nil {
			log.Fatal("Error verifying upload. Error: " + err.Error())
		}

		// Phase 3. Verify that the metadata for the uploaded data is commited to the ledger.
		castedTx := tmtypes.Tx(byteArrTranc)
		fileName := args[1];
		var fileDescription string;
		var blockHeight int64;
		if len(args) > 1 {
			fileDescription = args[2]
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
					fmt.Println("File successfully uploaded. CID: ", fileHash)
					fmt.Printf("Block height: %v\n", evt.Block.Height)
					blockHeight = evt.Block.Height
				}
			case <-time.After(time.Duration(conf.ClientConfig().NewBlockTimeout) * time.Second):
				fmt.Println("File was uploaded, but could not verify the ledger within the timeout. " +
					"Try running a status query with CID: " + fileHash)
		}

		// Write the metadata even if a timeout occured.
		types.WriteMetadata(fileHash, &types.MetadataEntry{Name:fileName, Description:fileDescription,
			StorageSample: *storageSample, Blockheight:blockHeight})
	},
}

func init() {
	dataCmd.AddCommand(uploadCmd)
}