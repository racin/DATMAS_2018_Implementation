package rpc

import (
	"github.com/racin/DATMAS_2018_Implementation/types"
	"bytes"
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"encoding/json"
	"io"
	"strings"
	"net/http"
	"io/ioutil"
)

func QueryIPFSproxy(httpClient *http.Client, rawProxyAddr string, ipfsproxy string, endpoint string,
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
		payload, contentType = GetMultipartValues(data)
	default:
		res.AddMessageAndError("Input must be of type *crypto.SignedStruct or *map[string]io.Reader.", types.CodeType_InternalError)
		return res
	}

	ipfsAddr := strings.Replace(rawProxyAddr, "$IpfsNode", ipfsproxy, 1)
	if response, err := httpClient.Post(ipfsAddr + endpoint, contentType, payload); err == nil{
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
