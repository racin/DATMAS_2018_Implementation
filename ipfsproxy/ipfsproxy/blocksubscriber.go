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
)
func (proxy *Proxy) SubToNewBlock(newBlock chan interface{}) error {
	return proxy.TMClient.Subscribe(context.Background(), "bcfs-ipfsproxy", tmtypes.EventQueryNewBlock, newBlock)
}

func (proxy *Proxy) setupAPI() error{
	// Get Tendermint blockchain API
	fmt.Printf("%+v\n", conf.ClientConfig().TendermintNodes)
	for _, ident := range conf.ClientConfig().TendermintNodes {
		apiAddr := strings.Replace(conf.ClientConfig().RemoteAddr, "$TmNode", proxy.GetAccessList().GetAddress(ident), 1)


		fmt.Println("Trying to connect to (TM_api: " + apiAddr)
		proxy.TMClient = rpcClient.NewHTTP(apiAddr, conf.ClientConfig().WebsocketEndPoint)
		if _, err := proxy.TMClient.Status(); err == nil {
			//conf.ClientConfig().RemoteAddr = apiAddr
			err := proxy.TMClient.Start()
			if err != nil {
				return err
			}
			proxy.TMIdent = ident
			return nil
		}
	}
	return errors.New("Fatal: Could not estabilsh connection with Tendermint API.")
}