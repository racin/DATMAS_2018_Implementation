package app

import (
	"os/user"
	"io/ioutil"
	"encoding/json"
)

const (
	Anonymous 	= 0
	User 		= 1
	Storage		= 3
	Consensus   = 4
)
const (
	listPath = "/.bcfs/accessList"
	listPathTest = "accessControl_test"
)
type accessList struct {
	Identities map[string]int `json:"identities"`
}

var acl accessList
func GetAccessList(params ...bool) (*accessList){
	lenP := len(params)
	if (lenP == 0 || (lenP > 0 && params[0] == false)) &&
		acl.Identities != nil {
			return &acl
	}

	usr, err := user.Current()
	if err != nil {
		panic("Could not get current user")
	}

	var path string
	if lenP > 1 && params[1] == true {
		path = listPathTest
	} else {
		path = usr.HomeDir + listPath
	}

	var z accessList = accessList{Identities:make(map[string]int)}
	if data, err := ioutil.ReadFile(path); err == nil {
		json.Unmarshal(data, &z)
	}
	if path == listPath {
		acl = z
	}

	return &z
}