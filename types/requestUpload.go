package types

import (
	"fmt"
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"encoding/base64"
)

func (ru *RequestUpload) CompareTo(other *RequestUpload) bool {
	return ru.Cid == other.Cid && ru.Length == other.Length && ru.IpfsNode == other.IpfsNode
}

func GetSignedRequestUploadFromMap(derivedStruct map[string]interface{}) *crypto.SignedStruct {
	if base, ok := derivedStruct["Base"]; ok {
		fmt.Println("dS: base")
		if signature, ok := derivedStruct["signature"]; ok {
			fmt.Println("dS: sig")
			if data, err := base64.StdEncoding.DecodeString(signature.(string)); err == nil {
				ss := &crypto.SignedStruct{Base: *GetRequestUploadFromMap(base.(map[string]interface{})), Signature: data}
				fmt.Printf("SS: %+v\n", ss)
				return ss
			}
		}
	}
	return nil
}

func GetRequestUploadFromMap(derivedStruct map[string]interface{}) *RequestUpload {
	if cid, ok := derivedStruct["cid"]; ok {
		fmt.Println("dS: a")
		if ipfsNode, ok := derivedStruct["ipfsNode"]; ok {
			fmt.Println("dS: b")
			if length, ok := derivedStruct["length"]; ok {
				fmt.Println("dS: c")
				return &RequestUpload{Cid: cid.(string), IpfsNode: ipfsNode.(string), Length:int64(length.(float64))}
			}
		}
	}
	return nil
}
