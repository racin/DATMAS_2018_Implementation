package app

import (
	"io/ioutil"
	"encoding/json"
	"os/user"
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
type Identity struct {
	AccessLevel 	int		`json:"level"`
	Name			string	`json:"name"`
	KeyPath			string	`json:"keypath"`
}
type accessList struct {
	Identities map[string]Identity `json:"identities"`
}

func GetAccessList(test ...bool) (*accessList){
	var path string
	usr, err := user.Current()
	if err != nil {
		panic("Could not get current user")
	}

	if (len(test) > 0 && test[0]) {
		path = listPathTest
	} else {
		path = usr.HomeDir + listPath
	}

	var z accessList = accessList{Identities:make(map[string]Identity)}

	if data, err := ioutil.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &z); err != nil {
			panic(err.Error())
		}
	}

	return &z
}

func WriteAccessList(acl *accessList){
	return // Do not use this function.
	if data, err := json.Marshal(&acl); err == nil {
		ioutil.WriteFile(listPath, data, 0600)
	}
}

/*
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
}*/