package app

import (
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	"strings"
	"fmt"
	"io/ioutil"
	"github.com/pkg/errors"
	rpcClient "github.com/tendermint/tendermint/rpc/client"
	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tendermint/rpc/core/types"
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"github.com/racin/DATMAS_2018_Implementation/types"
	"encoding/json"
	"bytes"
	"mime/multipart"
	"os"
	"io"
)


func (app *Application) setupTMRpcClients() {
	app.TMRpcClients = make(map[string]rpcClient.Client)
	for _, addr := range conf.AppConfig().TendermintNodes {
		if _, ok := app.TMRpcClients[addr]; ok {
			continue;
		}
		apiAddr := strings.Replace(conf.AppConfig().WebsocketAddr, "$TmNode", addr, 1)
		app.TMRpcClients[addr] = rpcClient.NewHTTP(apiAddr, conf.AppConfig().WebsocketEndPoint)
	}
}

func (app *Application) getIPFSProxyAddr() (string, error) {
	maxRetry := 10
	retries := 0
	RETRY_LOOP:
	// Get IPFS Proxy API
	for _, addr := range conf.AppConfig().IpfsNodes {
		ipfsAddr := strings.Replace(conf.AppConfig().IpfsProxyAddr, "$IpfsNode", addr, 1)
		fmt.Println("Trying to connect to (IPFS addr): " + ipfsAddr)

		if response, err := app.IpfsHttpClient.Get(ipfsAddr + conf.ClientConfig().IpfsIsupEndpoint); err != nil {
			if dat, err := ioutil.ReadAll(response.Body); err == nil && string(dat) == "true" {
				return ipfsAddr, nil
			}
		}
	}

	if retries++; retries != maxRetry {
		goto RETRY_LOOP
	}

	return "", errors.New("Fatal: Could not connect to IPFS Proxy")
}

func (app *Application) queryIPFSproxy(ipfsproxy string, endpoint string,
	input interface{}) (*types.IPFSReponse) {
	var payload *bytes.Buffer
	var contentType string
	res := &types.IPFSReponse{Codetype:types.CodeType_InternalError}
	switch data := input.(type){
		case *crypto.SignedStruct:
			if byteArr, err := json.Marshal(data); err != nil {
				res.Message = err.Error()
				return res
			} else {
				*payload = *bytes.NewBuffer(byteArr)
			}
			contentType = "application/json"
		case *map[string]io.Reader:
			payload, contentType = GetMultipartValues(data)
		default:
			res.Message = "Input must be of type *crypto.SignedStruct or *map[string]io.Reader."
			return res
	}

	ipfsAddr := strings.Replace(conf.AppConfig().IpfsProxyAddr, "$IpfsNode", ipfsproxy, 1)
	fmt.Println("Trying to connect to (IPFS addr): " + ipfsAddr)
	if response, err := app.IpfsHttpClient.Post(ipfsAddr + endpoint, contentType, payload); err == nil{
		if dat, err := ioutil.ReadAll(response.Body); err == nil{
			if err := json.Unmarshal(dat, res); err != nil {
				res.Message = err.Error()
			}
		} else {
			res.Message = err.Error()
		}
	} else {
		res.Message = err.Error()
	}

	return res
}

func GetMultipartValues(values *map[string]io.Reader) (buffer *bytes.Buffer, boundary string){
	var b bytes.Buffer
	var err error
	w := multipart.NewWriter(&b)
	defer w.Close()

	for index, element := range *values {
		var writer io.Writer
		// If file has a close method.
		if file, ok := element.(io.Closer); ok {
			defer file.Close()
		}

		// Check if a file is added. Else add it as a regular data element.
		if file, ok := element.(*os.File); ok {
			writer, err = w.CreateFormFile(index, file.Name());
		} else {
			writer, err = w.CreateFormField(index);
		}

		// If there was no errors creating the form element, try to copy the element to it
		if err == nil {
			io.Copy(writer, element)
		}
	}

	return &b, w.FormDataContentType()
}

type QueryBroadcastReponse struct {
	Identity		string
	Result			*core_types.ResultABCIQuery
	Err				error
}
func (app *Application) broadcastQuery(path string, data *[]byte, outChan chan<-*QueryBroadcastReponse){
	for key, value := range app.TMRpcClients {
		go func() {
			result, err := value.ABCIQuery(path, *data)
			outChan <- &QueryBroadcastReponse{Result: result, Err: err, Identity: key}
		}()

	}
}

func (app *Application) multicastQuery(path string, data cmn.HexBytes, tmNodes []string) map[string]*QueryBroadcastReponse{
	response := make(map[string]*QueryBroadcastReponse)
	for _, addr := range tmNodes {
		if tmClient, ok := app.TMRpcClients[addr]; !ok {
			continue // Not connected to node with that address
		} else {
			result, err := tmClient.ABCIQuery(path, data)
			response[addr] = &QueryBroadcastReponse{Result: result, Err: err}
		}
	}

	return response
}