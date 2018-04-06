package configuration

import (
	"os/user"
	"io/ioutil"
	"encoding/json"
)

const (
	appConf 	= "/.bcfs/appConfig"
	clientConf	= "/.bcfs/clientConfig"
)
var appConfig AppConfiguration
type AppConfiguration struct {
	ListenAddr 		string		`json:"listenAddr"`
	UploadAddr 		string		`json:"uploadAddr"`
	RpcType 		string		`json:"rpcType"`
	Info			string		`json:"appInfo"`
}

var clientConfig ClientConfiguration
type ClientConfiguration struct {
	EndPoint		string		`json:"endPoint"`
}

func LoadAppConfig(path ...string) (*AppConfiguration, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	var filePath string
	if len(path) > 0 {
		filePath = path[0]
	} else {
		filePath = usr.HomeDir + appConf
	}
	conf, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(conf, &appConfig); err != nil {
		return nil, err
	}

	return &appConfig, nil
}
func AppConfig() *AppConfiguration {
	return &appConfig
}

func LoadClientConfig(path ...string) (*ClientConfiguration, error){
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	var filePath string
	if len(path) > 0 {
		filePath = path[0]
	} else {
		filePath = usr.HomeDir + clientConf
	}
	conf, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(conf, &clientConfig); err != nil {
		return nil, err
	}

	return &clientConfig, nil
}
func ClientConfig() *ClientConfiguration{
	return &clientConfig
}