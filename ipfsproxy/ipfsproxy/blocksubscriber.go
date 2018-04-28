package ipfsproxy

import (
	rpcClient "github.com/tendermint/tendermint/rpc/client"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	"context"
	tmtypes "github.com/tendermint/tendermint/types"
	"strings"
	"fmt"
	"errors"
	"io/ioutil"
	"encoding/binary"
	"github.com/racin/DATMAS_2018_Implementation/types"
	"strconv"
)

func loadMaxSeenBlockHeight() int64 {
	if byteArr, err := ioutil.ReadFile(conf.IPFSProxyConfig().LastSeenBlockHeight); err == nil {
		blockHeight, _ := binary.Varint(byteArr)
		return blockHeight
	}
	return 0
}
func saveMaxSeenBlockHeight(height int64) error {
	byteArr := make([]byte, binary.MaxVarintLen64)
	_ = binary.PutVarint(byteArr, height)
	if err := ioutil.WriteFile(conf.IPFSProxyConfig().LastSeenBlockHeight, byteArr, 0600); err != nil {
		return err
	}
	return nil
}
func (proxy *Proxy) SubscribeToNewBlocks() {
	newBlockCh := make(chan interface{}, 1)
	proxy.subToNewBlock(newBlockCh)
	for {
		select {
		case b := <-newBlockCh:
			fmt.Println("New block!")
			evt, ok := b.(tmtypes.EventDataNewBlock)
			if !ok {
				// Consensus node shut down
				proxy.subToNewBlock(newBlockCh)
				break
			}
			if proxy.handleBlock(evt.Block) == types.CodeType_BCFSInvalidBlockHeight {
				proxy.processNewBlocks(evt.Block.Height)
			}
		}
	}
}

func (proxy *Proxy) processNewBlocks(height int64) error {
	if !proxy.TMClient.IsRunning() {
		if err := proxy.setupTMConnection(); err != nil {
			return err
		}
	}
	for i:=loadMaxSeenBlockHeight()+1; i<height; i++ {
		if block, err := proxy.TMClient.Block(&i); err != nil {
			return err
		} else if codetype := proxy.handleBlock(block.Block); codetype != types.CodeType_OK {
			return errors.New("Error handling block. Type: " + string(codetype))
		}
	}
	return nil
}

func (proxy *Proxy) handleBlock(block *tmtypes.Block) types.CodeType{
	fmt.Println("Handling block height: " + strconv.Itoa(int(block.Height)))
	if block == nil || block.ValidateBasic() != nil {
		return types.CodeType_BCFSInvalidBlock// Could not validate the block. Do not process it.
	}
	seenHeight := loadMaxSeenBlockHeight()
	if seenHeight+1 != block.Height {
		return types.CodeType_BCFSInvalidBlockHeight
	}
	for i := int64(0); i < block.NumTxs; i++ {
		_, tx, _ := types.UnmarshalTransaction([]byte(block.Txs[i].String()))
		// Attempt to PIN all new upload transactions
		if ipfsResp, ok := tx.Data.(*types.RequestUpload); ok {
			if proxy.fingerprint != ipfsResp.IpfsNode {
				continue
			}
			fmt.Println("Pinning file with CID: " + ipfsResp.Cid)
			proxy.pinFile(ipfsResp.Cid)
		}
	}
	fmt.Println("Updating seen block height to: " + strconv.Itoa(int(block.Height)))
	saveMaxSeenBlockHeight(block.Height)
	return types.CodeType_OK
}

func (proxy *Proxy) subToNewBlock(newBlockCh chan interface{}) error {
	if err := proxy.setupTMConnection(); err != nil {
		return err
	}
	if err := proxy.TMClient.Subscribe(context.Background(), "bcfs-ipfsproxy", tmtypes.EventQueryNewBlock, newBlockCh); err != nil {
		return err
	}
	return nil
}
func (proxy *Proxy) setupTMConnection() error{
	// Get Tendermint blockchain API
	fmt.Printf("%+v\n", conf.ClientConfig().TendermintNodes)
	for _, ident := range conf.ClientConfig().TendermintNodes {
		apiAddr := strings.Replace(conf.IPFSProxyConfig().WebsocketAddr, "$TmNode", proxy.GetAccessList().GetAddress(ident), 1)
		fmt.Println("Trying to connect to (TM_api: " + apiAddr)
		proxy.TMClient = rpcClient.NewHTTP(apiAddr, conf.IPFSProxyConfig().WebsocketEndPoint)
		if _, err := proxy.TMClient.Status(); err == nil {
			err := proxy.TMClient.Start()
			if err != nil {
				continue // Could not start subscription event at this node for some reason. Try another one.
			}
			proxy.TMIdent = ident
			return nil
		}
	}
	return errors.New("Fatal: Could not estabilsh connection with Tendermint API.")
}