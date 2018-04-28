package ipfsproxy

import (
	rpcClient "github.com/tendermint/tendermint/rpc/client"
	"github.com/racin/DATMAS_2018_Implementation/rpc"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	"context"
	tmtypes "github.com/tendermint/tendermint/types"
	"strings"
	"fmt"
	"errors"
	"time"
	"io/ioutil"
	"encoding/binary"
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
			evt, ok := b.(tmtypes.EventDataNewBlock)
			if !ok {
				// Consensus node shut down
				proxy.subToNewBlock(newBlockCh)
			}
			// Validate
			if err := evt.Block.ValidateBasic(); err != nil {
				// Could not validate this block. Do nothing.
				break
			}
			proxy.handleValidBlock(evt.Block)
		}
	}
}

func (proxy *Proxy) handleValidBlock(block *tmtypes.Block) {
	block.Height
}

func (proxy *Proxy) subToNewBlock(newBlockCh chan interface{}) error {
	if err := proxy.setupAPI(); err != nil {
		return err
	}
	if err := proxy.TMClient.Subscribe(context.Background(), "bcfs-ipfsproxy", tmtypes.EventQueryNewBlock, newBlockCh); err != nil {
		return err
	}
}
func (proxy *Proxy) setupAPI() error{
	// Get Tendermint blockchain API
	fmt.Printf("%+v\n", conf.ClientConfig().TendermintNodes)
	for _, ident := range conf.ClientConfig().TendermintNodes {
		apiAddr := strings.Replace(conf.ClientConfig().RemoteAddr, "$TmNode", proxy.GetAccessList().GetAddress(ident), 1)


		fmt.Println("Trying to connect to (TM_api: " + apiAddr)
		proxy.TMClient = rpcClient.NewHTTP(apiAddr, conf.ClientConfig().WebsocketEndPoint)
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