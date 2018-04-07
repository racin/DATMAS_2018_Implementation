package configuration

import (
	"os/user"
	"io/ioutil"
	"encoding/json"
	"strings"
)

const (
	appConf 		= "/.bcfs/appConfig"
	clientConf		= "/.bcfs/clientConfig"
)
var appConfig AppConfiguration
type AppConfiguration struct {
	BasePath 			string		`json:"basePath"`
	ListenAddr 			string		`json:"listenAddr"`
	UploadAddr 			string		`json:"uploadAddr"`
	UploadEndpoint		string		`json:"uploadEndPoint"`
	RpcType 			string		`json:"rpcType"`
	Info				string		`json:"appInfo"`
	PrivateKey			string		`json:"privateKey"`
	PublicKeys			string		`json:"publicKeys"`
	AccessList			string		`json:"accessList"`
}

var clientConfig ClientConfiguration
type ClientConfiguration struct {
	BasePath 				string		`json:"basePath"`
	RemoteAddr				string		`json:"remoteAddr"`
	UploadAddr				string		`json:"uploadAddr"`
	TendermintNodes			[]string	`json:"tendermintNodes"`
	WebsocketEndPoint		string		`json:"websocketEndPoint"`
	UploadEndPoint			string		`json:"uploadEndPoint"`
	PrivateKey				string		`json:"privateKey"`
	PublicKeys				string		`json:"publicKeys"`
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

	if strings.Contains(appConfig.BasePath,"$HOME") {
		appConfig.BasePath = strings.Replace(appConfig.BasePath, "$HOME", usr.HomeDir, 1)
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

	if strings.Contains(clientConfig.BasePath,"$HOME") {
		clientConfig.BasePath = strings.Replace(clientConfig.BasePath, "$HOME", usr.HomeDir, 1)
	}

	return &clientConfig, nil
}
func ClientConfig() *ClientConfiguration{
	return &clientConfig
}