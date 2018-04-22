package app

import (
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	"github.com/racin/DATMAS_2018_Implementation/rpc"
	"strings"
	"fmt"
	"io/ioutil"
	"github.com/pkg/errors"
	rpcClient "github.com/tendermint/tendermint/rpc/client"
	"github.com/tendermint/tendermint/rpc/core/types"
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"github.com/racin/DATMAS_2018_Implementation/types"
	"encoding/json"
	"bytes"

	"io"
)


func (app *Application) setupTMRpcClients() {
	app.TMRpcClients = make(map[string]rpcClient.Client)
	for _, ident := range conf.AppConfig().TendermintNodes {
		if _, ok := app.TMRpcClients[ident]; ok {
			continue;
		}
		addr := app.GetAccessList().GetAddress(ident)
		apiAddr := strings.Replace(conf.AppConfig().WebsocketAddr, "$TmNode", addr, 1)
		app.TMRpcClients[addr] = rpcClient.NewHTTP(apiAddr, conf.AppConfig().WebsocketEndPoint)
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

func (app *Application) queryIPFSproxy(ipfsproxy string, endpoint string,
	input interface{}) (*types.IPFSReponse) {
	var payload *bytes.Buffer
	var contentType string
	res := &types.IPFSReponse{}
	switch data := input.(type){
		case *crypto.SignedStruct:
			if byteArr, err := json.Marshal(data); err != nil {
				res.AddMessageAndError(err.Error(), types.CodeType_InternalError)
				return res
			} else {
				payload = bytes.NewBuffer(byteArr)
			}
			contentType = "application/json"
		case *map[string]io.Reader:
			payload, contentType = rpc.GetMultipartValues(data)
		default:
			res.AddMessageAndError("Input must be of type *crypto.SignedStruct or *map[string]io.Reader.", types.CodeType_InternalError)
			return res
	}

	fmt.Println("Was: " + conf.AppConfig().IpfsProxyAddr)
	ipfsAddr := strings.Replace(conf.AppConfig().IpfsProxyAddr, "$IpfsNode", ipfsproxy, 1)
	fmt.Println("Trying to connect to (IPFS addr): " + ipfsAddr)
	if response, err := app.IpfsHttpClient.Post(ipfsAddr + endpoint, contentType, payload); err == nil{
		if dat, err := ioutil.ReadAll(response.Body); err == nil{
			if err := json.Unmarshal(dat, res); err != nil {
				res.AddMessageAndError(err.Error(), types.CodeType_InternalError)
			}
		} else {
			res.AddMessageAndError(err.Error(), types.CodeType_InternalError)
		}
	} else {
		res.AddMessageAndError(err.Error(), types.CodeType_InternalError)
	}

	return res
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
	for _, addr := range tmNodes {
		if tmClient, ok := app.TMRpcClients[addr]; !ok {
			continue // Not connected to node with that address
		} else {
			result, err := tmClient.ABCIQuery(path, *data)
			response[addr] = &QueryBroadcastReponse{Result: result, Err: err}
		}
	}

	return response
}