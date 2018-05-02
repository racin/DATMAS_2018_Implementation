package client

import (
	"github.com/spf13/cobra"
	"log"
	"fmt"
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"github.com/racin/DATMAS_2018_Implementation/types"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	"io/ioutil"
	"github.com/racin/DATMAS_2018_Implementation/rpc"
)

// getAccountCmd represents the getAccount command
var dataGetCmd = &cobra.Command{
	Use:   "get [CID]",
	Short: "Download data",
	Long:  `Download data from the storage network.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			log.Fatal("Not enough arguments.")
		}
		// File and Name is required parameters.
		cid := args[0];
		tx := types.NewTx(cid, TheClient.fingerprint, types.TransactionType_DownloadData)
		stx, err := crypto.SignStruct(tx, TheClient.privKey)
		if err != nil {
			log.Fatal(err.Error())
		}

		ipfsResp := rpc.QueryIPFSproxy(TheClient.IPFSClient, conf.ClientConfig().IpfsProxyAddr,
			TheClient.GetAccessList().GetAddress(TheClient.IPFSIdent), conf.ClientConfig().IpfsGetEndpoint, stx)

		if ipfsResp.Codetype != types.CodeType_OK {
			log.Fatal("Could not download file. Error: " + string(ipfsResp.Message))
		}

		// Is the hash of the file the same as we requested?
		if ipfsHash, err := crypto.IPFSHashData(ipfsResp.Message); err != nil {
			log.Fatal("Error hashing file: %v", err)
		} else if ipfsHash != cid {
			log.Fatal("Hash of downloaded file was unexpected. Wanted: %v, Got: %v", cid, ipfsHash)
		}

		// See if we have a filename stored for this file.
		var filename string
		if me := types.GetMetadata(cid); me != nil {
			filename = me.Name
		} else {
			filename = cid
		}

		fullPath := conf.ClientConfig().BasePath + conf.ClientConfig().Downloads + filename
		if err := ioutil.WriteFile(fullPath, ipfsResp.Message, 0644); err != nil {
			log.Fatal(err.Error())
		} else {
			fmt.Printf("File successfully downloaded to: %v\n", fullPath)
		}
	},
}

func init() {
	dataCmd.AddCommand(dataGetCmd)
}