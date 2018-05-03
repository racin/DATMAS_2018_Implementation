package types

import (
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"encoding/base64"
)

func (ru *RequestUpload) CompareTo(other *RequestUpload) bool {
	return ru.Cid == other.Cid && ru.Length == other.Length && ru.IpfsNode == other.IpfsNode
}

func GetSignedRequestUploadFromMap(derivedStruct map[string]interface{}) *crypto.SignedStruct {
	if base, ok := derivedStruct["Base"]; ok {
		if signature, ok := derivedStruct["signature"]; ok {
			if data, err := base64.StdEncoding.DecodeString(signature.(string)); err == nil {
				ss := &crypto.SignedStruct{Base: GetRequestUploadFromMap(base.(map[string]interface{})), Signature: data}
				return ss
			}
		}
	}
	return nil
}

func GetRequestUploadFromMap(derivedStruct map[string]interface{}) *RequestUpload {
	if cid, ok := derivedStruct["cid"]; ok {
		if ipfsNode, ok := derivedStruct["ipfsNode"]; ok {
			if length, ok := derivedStruct["length"]; ok {
				return &RequestUpload{Cid: cid.(string), IpfsNode: ipfsNode.(string), Length:int64(length.(float64))}
			}
		}
	}
	return nil
}
