package client

import (
	"github.com/spf13/cobra"
	"log"
	"strings"
	"strconv"
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	"crypto/rand"
	"math/big"
	"math"
	"encoding/json"
	tmtypes "github.com/tendermint/tendermint/types"
	"time"
	"fmt"
	"github.com/racin/DATMAS_2018_Implementation/types"
)

var challengeCmd = &cobra.Command{
	Use:     "challenge [CID] [challenge]",
	Aliases: []string{"challenge"},
	Short:   "Challenge storage nodes",
	Long:    `Challenge storage nodes to prove that they still possess all the data for a CID.`,
	Run: func(cmd *cobra.Command, args []string) {
		var cid string
		var challengeIndices []uint64
		if len(args) < 1 {
			log.Fatal("Not enough arguments.")
		} else if len(args) == 2 {
			strArr := strings.Split(args[1], ",")
			for i, val := range strArr{
				if index, err := strconv.Atoi(val); err == nil {
					challengeIndices[i] = uint64(index)
				}
			}
		}
		var challenge *crypto.SignedStruct
		if challengeIndices == nil{
			me := GetMetadata(cid)
			if me == nil {
				log.Fatal("Could not find stored metadata for CID: " + cid)
			}
			challenge, _ = me.GenerateChallenge(TheClient.privKey)
		} else {
			nonce, err := rand.Int(rand.Reader, new(big.Int).SetUint64(math.MaxUint64)) // 1 << 64 - 1
			if err != nil {
				log.Fatal(err.Error()) // Could not generate nonce.
			}
			chal := &crypto.StorageChallenge{Identity:TheClient.fingerprint, Cid:cid, Challenge:challengeIndices, Nonce:nonce.Uint64()}
			if challenge, err = crypto.SignStruct(chal, TheClient.privKey); err != nil {
				log.Fatal(err.Error())
			}
		}
		byteArr, err := json.Marshal(challenge);
		if err != nil {
			log.Fatal(err.Error())
		}

		queryResp, err := TheClient.TMClient.ABCIQuery("/challenge", byteArr)
		if err != nil {
			log.Fatal(err.Error())
		}
		newBlockCh := make(chan interface{}, 1)
		if err := TheClient.SubToNewBlock(newBlockCh); err != nil {
			log.Fatal("Could not subscribe to new block events. Error: ", err.Error())
		}
		castedTx := tmtypes.Tx(byteArr)
		for {
			select {
			case b := <-newBlockCh:
				evt := b.(tmtypes.EventDataNewBlock)
				fmt.Printf("New block: %+v\n", evt.Block)
				// Validate
				if err := evt.Block.ValidateBasic(); err != nil {
					// System is broken. Notify administrators
					log.Fatal("Could not validate latest block. Error: ", err.Error())
				}
				for i := int64(0); i < evt.Block.NumTxs; i++ {
					fmt.Println("Trying to unmarshal Tx")
					// Check if the transaction contains a StorageProofCollection
					if _, tx, err := types.UnmarshalTransaction([]byte(evt.Block.Txs[i])); err == nil {
						// Attempt to PIN all new upload transactions
						if ipfsResp, ok := tx.Data.(*crypto.SignedStruct).Base.(*types.RequestUpload); ok {
							if proxy.fingerprint != ipfsResp.IpfsNode {
								continue
							}
							fmt.Println("Pinning file with CID: " + ipfsResp.Cid)
							proxy.pinFile(ipfsResp.Cid)
						}
					}
				}
				if evt.Block.Txs.Index(castedTx) > -1 {
					// Transaction is put in the latest block.
					fmt.Println("File successfully uploaded. CID: ", fileHash)
					fmt.Printf("Block height: %v\n", evt.Block.Height)
					WriteMetadata(fileHash, &MetadataEntry{Name: fileName, Description: fileDescription,
						StorageSample: *storageSample, Blockheight: evt.Block.Height})
				}
			case <-time.After(time.Duration(conf.ClientConfig().NewBlockTimeout) * time.Second):
				fmt.Println("File was uploaded, but could not verify the ledger within the timeout. " +
					"Try running a status query with CID: " + fileHash)
				return
			}
		}

	},
}

func init() {
	RootCmd.AddCommand(challengeCmd)
}