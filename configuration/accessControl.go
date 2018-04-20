package configuration

import (
	"io/ioutil"
	"encoding/json"
)

type Context int
type NodeType int
const (
	Anonymous 	NodeType = iota
	Client 		NodeType = 1
	Storage		NodeType = 2
	Consensus   NodeType = 3

	app 	Context = iota
	ipfs	Context = 1
	test	Context = 2

	ListPathTest = "test/accessControl_test"
)

type Identity struct {
	Type 		NodeType    	`json:"type"`
	Name        string 			`json:"name"`
	PublicKey   string 			`json:"publickey"`
	Address  	string 			`json:"address"`
}
type AccessList struct {
	Identities map[string]*Identity `json:"identities"`
}

var z map[string]*AccessList

func init() {
	z = make(map[string]*AccessList)
}
func GetAccessList(path string) (*AccessList){
	if val, ok := z[path]; ok {
		return val
	}

	z[path] = &AccessList{Identities:make(map[string]*Identity)}
	if data, err := ioutil.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, z[path]); err != nil {
			panic(err.Error())
		}
	} else {
		panic(err.Error())
	}

	return z[path]
}

func WriteAccessList(acl *AccessList, path string){
	return // Do not use this function.
	if data, err := json.Marshal(acl); err == nil {
		ioutil.WriteFile(path, data, 0600)
	}
	z[path] = acl // Force reloading of access list.
}

func (acl *AccessList) GetAddress(id string) string {
	if ident, ok := acl.Identities[id]; ok {
		return ident.Address
	}
	return "";
}