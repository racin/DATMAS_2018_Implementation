package app

import (
	"io/ioutil"
	"encoding/json"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	"fmt"
)

const (
	Anonymous 	= 0
	User 		= 1
	Storage		= 3
	Consensus   = 4
)
const (
	listPathTest = "accessControl_test"
)
type Identity struct {
	AccessLevel int    `json:"level"`
	Name        string `json:"name"`
	PublicKey   string `json:"publickey"`
}
type accessList struct {
	Identities map[string]Identity `json:"identities"`
}

func GetAccessList(test ...bool) (*accessList){
	var path string
	fmt.Printf("%+v", conf.AppConfig())
	if len(test) > 0 && test[0] {
		path = listPathTest
	} else {
		path = conf.AppConfig().BasePath + conf.AppConfig().AccessList
	}

	var z accessList = accessList{Identities:make(map[string]Identity)}
	fmt.Println(path)
	if data, err := ioutil.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &z); err != nil {
			panic(err.Error())
		}
	} else {
		panic(err.Error())
	}

	return &z
}

func WriteAccessList(acl *accessList){
	return // Do not use this function.
	if data, err := json.Marshal(&acl); err == nil {
		ioutil.WriteFile(conf.AppConfig().BasePath + conf.AppConfig().AccessList, data, 0600)
	}
}