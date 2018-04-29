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
	ipfsProxyConf	= "/.bcfs/ipfsProxyConfig"
)

var appConfig AppConfiguration
type AppConfiguration struct {
	BasePath 						string		`json:"basePath"`
	ListenAddr 						string		`json:"listenAddr"`
	UploadAddr 						string		`json:"uploadAddr"`
	UploadEndpoint					string		`json:"uploadEndpoint"`
	RpcType 						string		`json:"rpcType"`
	Info							string		`json:"appInfo"`
	PrivateKey						string		`json:"privateKey"`
	PublicKeys						string		`json:"publicKeys"`
	AccessList						string		`json:"accessList"`
	StorageSamples					string		`json:"storageSamples"`
	TempUploadPath					string		`json:"tempUploadPath"`
	TendermintNodes					[]string	`json:"tendermintNodes"`
	TmQueryTimeoutSeconds			int			`json:"tmQueryTimeoutSeconds"`
	WebsocketEndPoint				string		`json:"websocketEndpoint"`
	WebsocketAddr					string		`json:"websocketAddr"`
	IpfsNodes						[]string	`json:"ipfsNodes"`
	IpfsProxyAddr					string		`json:"ipfsProxyAddr"`
	IpfsProxyTimeoutSeconds			int			`json:"ipfsProxyTimeoutSeconds"`
	IpfsIsupEndpoint				string		`json:"ipfsIsupEndpoint"`
	IpfsStatusallEndpoint			string		`json:"ipfsStatusallEndpoint"`
	IpfsPinfileEndpoint				string		`json:"ipfsPinfileEndpoint"`
	IpfsUnpinfileEndpoint			string		`json:"ipfsUnpinfileEndpoint"`
	IpfsGetEndpoint					string		`json:"ipfsGetEndpoint"`
	IpfsStatusEndpoint				string		`json:"ipfsStatusEndpoint"`
	IpfsAddnopinEndpoint			string		`json:"ipfsAddnopinEndpoint"`
	IpfsChallengeEndpoint			string		`json:"ipfsChallengeEndpoint"`
}

var clientConfig ClientConfiguration
type ClientConfiguration struct {
	BasePath 						string		`json:"basePath"`
	RemoteAddr						string		`json:"remoteAddr"`
	Metadata						string		`json:"metadata"`
	UploadAddr						string		`json:"uploadAddr"`
	TendermintNodes					[]string	`json:"tendermintNodes"`
	WebsocketEndPoint				string		`json:"websocketEndpoint"`
	UploadEndPoint					string		`json:"uploadEndpoint"`
	UploadTimeoutSeconds			int			`json:"uploadTimeoutSeconds"`
	PrivateKey						string		`json:"privateKey"`
	PublicKeys						string		`json:"publicKeys"`
	AccessList						string		`json:"accessList"`
	IpfsNodes						[]string	`json:"ipfsNodes"`
	IpfsProxyTimeoutSeconds			int			`json:"ipfsProxyTimeoutSeconds"`
	IpfsProxyAddr					string		`json:"ipfsProxyAddr"`
	IpfsIsupEndpoint				string		`json:"ipfsIsupEndpoint"`
	IpfsStatusallEndpoint			string		`json:"ipfsStatusallEndpoint"`
	IpfsPinfileEndpoint				string		`json:"ipfsPinfileEndpoint"`
	IpfsUnpinfileEndpoint			string		`json:"ipfsUnpinfileEndpoint"`
	IpfsGetEndpoint					string		`json:"ipfsGetEndpoint"`
	IpfsStatusEndpoint				string		`json:"ipfsStatusEndpoint"`
	IpfsAddnopinEndpoint			string		`json:"ipfsAddnopinEndpoint"`
	IpfsChallengeEndpoint			string		`json:"ipfsChallengeEndpoint"`
	NewBlockTimeout					int			`json:"newBlockTimeoutSeconds"`
}

var ipfsProxyConfig IPFSProxyConfiguration
type IPFSProxyConfiguration struct {
	BasePath 						string		`json:"basePath"`
	AccessList						string		`json:"accessList"`
	PrivateKey						string		`json:"privateKey"`
	PublicKeys						string		`json:"publicKeys"`
	LastSeenBlockHeight				string		`json:"lastSeenBlockHeight"`
	ListenAddr 						string		`json:"listenAddr"`
	TempUploadPath					string		`json:"tempUploadPath"`
	TendermintNodes					[]string	`json:"tendermintNodes"`
	TmQueryTimeoutSeconds			int			`json:"tmQueryTimeoutSeconds"`
	WebsocketEndPoint				string		`json:"websocketEndpoint"`
	WebsocketAddr					string		`json:"websocketAddr"`
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

func LoadIPFSProxyConfig(path ...string) (*IPFSProxyConfiguration, error){
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	var filePath string
	if len(path) > 0 {
		filePath = path[0]
	} else {
		filePath = usr.HomeDir + ipfsProxyConf
	}
	conf, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(conf, &ipfsProxyConfig); err != nil {
		return nil, err
	}

	if strings.Contains(ipfsProxyConfig.BasePath,"$HOME") {
		ipfsProxyConfig.BasePath = strings.Replace(ipfsProxyConfig.BasePath, "$HOME", usr.HomeDir, 1)
	}

	return &ipfsProxyConfig, nil
}
func IPFSProxyConfig() *IPFSProxyConfiguration{
	return &ipfsProxyConfig
}