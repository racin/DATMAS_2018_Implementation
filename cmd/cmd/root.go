package cmd

import (
	"fmt"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	"github.com/racin/DATMAS_2018_Implementation/client"
	"github.com/racin/DATMAS_2018_Implementation/app"
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"github.com/spf13/cobra"

	"os"
	"strings"
	"io"
	"math/rand"
	"time"
	"strconv"
	"github.com/racin/DATMAS_2018_Implementation/types"
)

var RootCmd = &cobra.Command{
	Use:   "bcfs",
	Short: "Block Chain File System",
	Long: `Implementation of Block Chain File System for Master Thesis in Computer Science at UiS 2018.
Written by Racin Nygaard.	`,
}

var cfgFile string
func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Config file (default is $HOME/.bcfs/clientConfig)")
}

func initConfig() {
	var err error
	if cfgFile != "" {
		_, err = conf.LoadClientConfig(cfgFile);
	} else {
		_, err = conf.LoadClientConfig();
	}
	if err != nil {
		fmt.Println("Could not load configuration:", err)
		os.Exit(1)
	}
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	fmt.Println("a")
	fmt.Println("Main client")
}

func getSignedTransaction(txtype app.TransactionType, data interface{}) (stranc *app.SignedTransaction) {
	keys, err := crypto.LoadPrivateKey(conf.ClientConfig().BasePath + conf.ClientConfig().PrivateKey)
	if err != nil {
		panic("Could not load private key. Use the --generateKeys option to generate a new one. Error: " + err.Error())
	}

	fp, err := crypto.GetFingerPrint(keys)
	if err != nil {
		panic("Could not load fingerprint of public key. Use the --generateKeys to generate a new one. Error: " + err.Error())
	}

	stranc, err = app.NewTx(data, fp, txtype).Sign(keys);
	if err != nil {
		panic("Could not sign transaction. Private/Public key pair may not match. Use the --generateKeys to generate a new one. Error: " + err.Error())
	}
	return
}

var rootAPI client.API
func getAPI() client.API {
	if rootAPI != nil {
		return rootAPI
	}
	tmApiFound, tmUplApiFound, ipfsProxyFound := false, false, false
	var remoteAddr string

	// Get Tendermint blockchain API
	s1 := rand.NewSource(time.Now().UnixNano())
	reqNum := strconv.Itoa(rand.New(s1).Int())
	fmt.Printf("%+v\n", conf.ClientConfig().TendermintNodes)
	for _, addr := range conf.ClientConfig().TendermintNodes {
		if !tmApiFound {
			rootAPI = client.NewTM_API(strings.Replace(conf.ClientConfig().RemoteAddr, "$TmNode", addr, 1))
			fmt.Println("Trying to connect to (TM_api: " + strings.Replace(conf.ClientConfig().RemoteAddr, "$TmNode", addr, 1))
			if _, err := rootAPI.GetBase().TM.Status(); err == nil {
				remoteAddr = addr
				tmApiFound = true
			}
		}

		if !tmUplApiFound {
			uploadAddr := strings.Replace(conf.ClientConfig().UploadAddr, "$TmNode", addr, 1)
			fmt.Println("Trying to connect to (TM_uplApi): " + uploadAddr)

			values := map[string]io.Reader{
				"Status":    strings.NewReader(reqNum),
			}

			response := rootAPI.GetBase().SendMultipartFormData(uploadAddr, &values);
			if response.Codetype == types.CodeType_OK  && response.Message == reqNum{
				remoteAddr = addr
				tmUplApiFound = true
			} else{
				fmt.Printf("Error Response: %+v\n", response)
			}
		}
	}


	if !apiOk {
		panic("Fatal: Could not estabilsh connection with API.")
	}
	apiOk = false
	conf.ClientConfig().RemoteAddr = strings.Replace(conf.ClientConfig().RemoteAddr, "$TmNode", remoteAddr, 1)

	// Get Tendermint Upload API
	for _, addr := range conf.ClientConfig().TendermintNodes {

	}

	//NewUploadHTTPClient
	conf.ClientConfig().UploadAddr = strings.Replace(conf.ClientConfig().UploadAddr, "$TmNode", remoteAddr, 1)
	if !apiOk {
		panic("Fatal: Could not estabilsh connection with API.")
	}

	apiOk = false
	conf.ClientConfig().RemoteAddr = strings.Replace(conf.ClientConfig().RemoteAddr, "$TmNode", remoteAddr, 1)
	for _, addr := range conf.ClientConfig().TendermintNodes {
		uploadAddr := strings.Replace(conf.ClientConfig().UploadAddr, "$TmNode", addr, 1)
		fmt.Println("Trying to connect to: " + uploadAddr)
		reqNum := strconv.Itoa(rand.New(s1).Int())
		values := map[string]io.Reader{
			"Status":    strings.NewReader(reqNum),
		}

		response := rootAPI.GetBase().SendMultipartFormData(uploadAddr, &values);
		if response.Codetype == types.CodeType_OK  && response.Message == reqNum{
			remoteAddr = addr
			apiOk = true
			break
		} else{
			fmt.Printf("Error Response: %+v\n", response)
		}
	}

	//NewUploadHTTPClient
	conf.ClientConfig().UploadAddr = strings.Replace(conf.ClientConfig().UploadAddr, "$TmNode", remoteAddr, 1)
	if !apiOk {
		panic("Fatal: Could not estabilsh connection with API.")
	}

	return rootAPI
}
