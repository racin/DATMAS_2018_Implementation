package configuration

import (
	"io/ioutil"
	"encoding/json"
)

type Context int
type AccessLevel int
const (
	Anonymous 	AccessLevel = iota
	User 		AccessLevel = 1
	Storage		AccessLevel = 2
	Consensus   AccessLevel = 3

	app 	Context = iota
	ipfs	Context = 1
	test	Context = 2

	listPathTest = "accessControl_test"
)

type Identity struct {
	AccessLevel AccessLevel    	`json:"level"`
	Name        string 			`json:"name"`
	PublicKey   string 			`json:"publickey"`
}
type AccessList struct {
	Identities map[string]Identity `json:"identities"`
}

func GetAccessList(path string) (*AccessList){
	var z AccessList = AccessList{Identities:make(map[string]Identity)}

	if data, err := ioutil.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &z); err != nil {
			panic(err.Error())
		}
	} else {
		panic(err.Error())
	}

	return &z
}

func WriteAccessList(acl *AccessList, path string){
	return // Do not use this function.
	if data, err := json.Marshal(acl); err == nil {
		ioutil.WriteFile(path, data, 0600)
	}
}