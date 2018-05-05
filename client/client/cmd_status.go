package client

import (
	"github.com/spf13/cobra"
	"log"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	"fmt"
	"github.com/racin/DATMAS_2018_Implementation/types"
	"github.com/racin/DATMAS_2018_Implementation/rpc"
	"encoding/json"
	"github.com/ipfs/ipfs-cluster/api"
)

var statusCmd = &cobra.Command{
	Use:     "status [CID] [StorageNode]",
	Aliases: []string{"challenge"},
	Short:   "Issue a simple status check to a storage node",
	Long:    `Challenge storage nodes to prove that they still possess all the data for a CID.`,
	Run: func(cmd *cobra.Command, args []string) {
		var storageNode string
		if len(args) < 1 {
			log.Fatal("Not enough arguments.")
		} else if len(args) > 1 {
			storageNode = TheClient.GetAccessList().GetAddress(args[1])
		} else {
			storageNode = TheClient.GetAccessList().GetAddress(TheClient.IPFSIdent)
		}

		cid := args[0]
		stx := TheClient.GetSignedTransaction(types.TransactionType_IPFSStatus, cid)
		var ipfsResp *types.IPFSReponse
		switch cid {
		case "all":
			ipfsResp = rpc.QueryIPFSproxy(TheClient.IPFSClient, conf.ClientConfig().IpfsProxyAddr,
				storageNode, conf.ClientConfig().IpfsStatusallEndpoint, stx)
		default:
			ipfsResp = rpc.QueryIPFSproxy(TheClient.IPFSClient, conf.ClientConfig().IpfsProxyAddr,
				storageNode, conf.ClientConfig().IpfsStatusEndpoint, stx)
		}

		if ipfsResp.Codetype != types.CodeType_OK {
			log.Fatal("Could not get status from storage node.")
		}

		apiInfoArr := make([]api.GlobalPinInfo, 0)
		apiInfo := &api.GlobalPinInfo{}
		if err := json.Unmarshal(ipfsResp.Message, apiInfo); err == nil {
			for _, info := range apiInfo.PeerMap {
				var errStr string
				if info.Error != "" {
					errStr = ", Error: " + info.Error
				}
				fmt.Printf("Response from node %v:\n   CID: %v, Status: %v, Date: %v%v\n", storageNode, info.Cid, info.Status, info.TS, errStr)
				break
			}
		} else if err := json.Unmarshal(ipfsResp.Message, &apiInfoArr); err == nil {
			fmt.Printf("Response from node %v:\n", storageNode)
			for _, subInfo := range apiInfoArr {
				for _, info := range subInfo.PeerMap {
					var errStr string
					if info.Error != "" {
						errStr = ", Error: " + info.Error
					}
					fmt.Printf("   CID: %v, Status: %v, Date: %v%v\n", info.Cid, info.Status, info.TS, errStr)
					break
				}
			}
		} else {
			log.Fatal("Could not get status from storage node.")
		}
	},
}

func init() {
	RootCmd.AddCommand(statusCmd)
}