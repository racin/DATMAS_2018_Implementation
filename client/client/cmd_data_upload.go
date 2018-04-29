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
		fileStat, err := file.Stat()
		if err != nil {
			log.Fatal("Could not get Stat of file. Error: ", err.Error())
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


		// Phase 1. Upload data to one Storage node.
		ipfsIdentity, ipfsPubKey := crypto.GetIdentityPublicKey(TheClient.IPFSIdent, TheClient.GetAccessList(),
			conf.ClientConfig().BasePath + conf.ClientConfig().PublicKeys)
		fmt.Printf("Identity: %+v\n", ipfsIdentity)
		if ipfsIdentity == nil {
			log.Fatal("Could not find IPFS node in the access list.")
		}
		sentReqUpload := &types.RequestUpload{Cid:fileHash, IpfsNode:TheClient.IPFSIdent, Length:fileStat.Size()}
		stranc := TheClient.GetSignedTransaction(types.TransactionType_UploadData, sentReqUpload)
		fmt.Printf("Tranc: %+v\n", stranc.Base.(*types.Transaction))
		byteArr, _ := json.Marshal(stranc)
		values := map[string]io.Reader{
			"file":    file2,
			"transaction": bytes.NewReader(byteArr),
		}

		ipfsResponse := TheClient.UploadDataToIPFS(&values);
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
		z := &types.RequestUpload{Cid:ipfsStx.Base.(*types.RequestUpload).Cid, Length: ipfsStx.Base.(*types.RequestUpload).Length,
			IpfsNode: ipfsStx.Base.(*types.RequestUpload).IpfsNode}
		fmt.Printf("IpfsStx: %+v\n", ipfsStx)
		fmt.Printf("IpfsStx Base: %+v\n", z)
		fmt.Printf("IpfsStx Base2: %+v\n", ipfsStx.Base.(*types.RequestUpload))
		fmt.Printf("IpfsPubkey: %+v\n", ipfsPubKey)
		if !ipfsStx.Verify(ipfsPubKey) {
			log.Fatal("Could not verify IPFS signature.")
		}

		// Phase 1b. Generate a sample for the file
		storageSample := crypto.GenerateStorageSample(&fileBytes)

		// Phase 2. Send metadata to TM
		newBlockCh := make(chan interface{}, 1)
		if err := TheClient.SubToNewBlock(newBlockCh); err != nil {
			log.Fatal("Could not subscribe to new block events. Error: ", err.Error())
		}

		fmt.Printf("Getting signed tranc. \n")
		strancTM := TheClient.GetSignedTransaction(types.TransactionType_UploadData, ipfsStx)
		fmt.Printf("StrancTM: %+v\n", strancTM)
		fmt.Printf("Hash strancTM: %v\n", crypto.HashStruct(strancTM))
		fmt.Printf("Hash ipfsStx: %v\n", crypto.HashStruct(ipfsStx))
		if _, err := TheClient.VerifyUpload(strancTM); err != nil {
			log.Fatal("Error verifying upload. Error: " + err.Error())
		}

		// Phase 3. Verify the uploaded data is commited to the ledger
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
					fmt.Printf("Block height: %v\n", evt.Block.Height)
					WriteMetadata(fileHash, &MetadataEntry{Name:fileName, Description:fileDescription,
						StorageSample: *storageSample, Blockheight:evt.Block.Height})
				}
			case <-time.After(time.Duration(conf.ClientConfig().NewBlockTimeout) * time.Second):
				fmt.Println("File was uploaded, but could not verify the ledger within the timeout. " +
					"Try running a status query with CID: " + fileHash)
		}
	},
}

func init() {
	dataCmd.AddCommand(uploadCmd)
}