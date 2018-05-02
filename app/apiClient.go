package app

import (
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	"strings"
	"fmt"
	"io/ioutil"
	"github.com/pkg/errors"
	rpcClient "github.com/tendermint/tendermint/rpc/client"
	"github.com/tendermint/tendermint/rpc/core/types"
)


func (app *Application) setupTMRpcClients() {
	app.TMRpcClients = make(map[string]rpcClient.Client)
	for _, ident := range conf.AppConfig().TendermintNodes {
		if _, ok := app.TMRpcClients[ident]; ok {
			continue;
		}
		addr := app.GetAccessList().GetAddress(ident)
		apiAddr := strings.Replace(conf.AppConfig().WebsocketAddr, "$TmNode", addr, 1)
		app.TMRpcClients[ident] = rpcClient.NewHTTP(apiAddr, conf.AppConfig().WebsocketEndPoint)
	}
}

func (app *Application) getIPFSProxyAddr() (string, error) {
	maxRetry := 10
	retries := 0
	RETRY_LOOP:
	// Get IPFS Proxy API
	for _, ident := range conf.AppConfig().IpfsNodes {
		addr := app.GetAccessList().GetAddress(ident)
		ipfsAddr := strings.Replace(conf.AppConfig().IpfsProxyAddr, "$IpfsNode", addr, 1)
		fmt.Println("Trying to connect to (IPFS addr): " + ipfsAddr)

		if response, err := app.IpfsHttpClient.Get(ipfsAddr + conf.ClientConfig().IpfsIsupEndpoint); err != nil {
			if dat, err := ioutil.ReadAll(response.Body); err == nil && string(dat) == "true" {
				return ipfsAddr, nil
			}
		}
	}

	if retries++; retries < maxRetry {
		goto RETRY_LOOP
	}

	return "", errors.New("Fatal: Could not connect to IPFS Proxy")
}

type QueryBroadcastReponse struct {
	Identity		string
	Result			*core_types.ResultABCIQuery
	Err				error
}

func (app *Application) broadcastQuery(path string, data *[]byte, outChan chan<-*QueryBroadcastReponse, done chan struct{}){
	for key, value := range app.TMRpcClients {
		fmt.Println("Go func to: " + key)
		go func(k string, v rpcClient.Client) {
			result, err := v.ABCIQuery(path, *data)
			select {
				case <-done:
					return
				case outChan <- &QueryBroadcastReponse{Result: result, Err: err, Identity: k}:
			}
		}(key, value)
	}
}

func (app *Application) multicastQuery(path string, data *[]byte, tmNodes []string) map[string]*QueryBroadcastReponse{
	response := make(map[string]*QueryBroadcastReponse)
	for _, ident := range tmNodes {
		if tmClient, ok := app.TMRpcClients[ident]; !ok {
			continue // Not connected to node with that address
		} else {
			result, err := tmClient.ABCIQuery(path, *data)
			response[ident] = &QueryBroadcastReponse{Result: result, Err: err}
		}
	}

	return response
}