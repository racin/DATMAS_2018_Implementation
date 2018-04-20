package cmd

import (
	"fmt"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	"github.com/racin/DATMAS_2018_Implementation/client"
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"github.com/spf13/cobra"

	"os"
	"strings"
	"io"
	"math/rand"
	"time"
	"strconv"
	"github.com/racin/DATMAS_2018_Implementation/types"
	tmtypes "github.com/tendermint/tendermint/types"
	"io/ioutil"
	"context"
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

func subToNewBlock(newBlock chan interface{}) error {
	return getAPI().GetBase().TMClient.Subscribe(context.Background(), "bcfs-client", tmtypes.EventQueryNewBlock, newBlock)
}

func getSignedTransaction(txtype types.TransactionType, data interface{}) (stranc *crypto.SignedStruct) {
	keys, err := crypto.LoadPrivateKey(conf.ClientConfig().BasePath + conf.ClientConfig().PrivateKey)
	if err != nil {
		panic("Could not load private key. Use the --generateKeys option to generate a new one. Error: " + err.Error())
	}

	fp, err := crypto.GetFingerprint(keys)
	if err != nil {
		panic("Could not load fingerprint of public key. Use the --generateKeys to generate a new one. Error: " + err.Error())
	}

	stranc, err = crypto.SignStruct(types.NewTx(data, fp, txtype), keys);
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

	// Get Tendermint blockchain API
	s1 := rand.NewSource(time.Now().UnixNano())
	reqNum := strconv.Itoa(rand.New(s1).Int())
	fmt.Printf("%+v\n", conf.ClientConfig().TendermintNodes)
	for _, addr := range conf.ClientConfig().TendermintNodes {
		if !tmApiFound {
			apiAddr := strings.Replace(conf.ClientConfig().RemoteAddr, "$TmNode", addr, 1)
			rootAPI = client.NewTM_API(apiAddr)
			fmt.Println("Trying to connect to (TM_api: " + apiAddr)
			if _, err := rootAPI.GetBase().TMClient.Status(); err == nil {
				//conf.ClientConfig().RemoteAddr = apiAddr
				tmApiFound = true
			}
		}

		if !tmUplApiFound {
			uploadAddr := strings.Replace(conf.ClientConfig().UploadAddr, "$TmNode", addr, 1)
			fmt.Println("Trying to connect to (TM_uplApi): " + uploadAddr)

			values := map[string]io.Reader{
				"Status":    strings.NewReader(reqNum),
			}

			uploadAPI := uploadAddr + conf.ClientConfig().UploadEndPoint
			response := rootAPI.GetBase().SendMultipartFormData(uploadAPI, &values);
			if response.Codetype == types.CodeType_OK && response.Message == reqNum{
				rootAPI.GetBase().TMUploadAPI = uploadAPI
				//conf.ClientConfig().UploadAddr = uploadAddr
				tmUplApiFound = true
			}
		}
	}

	if !tmApiFound || !tmUplApiFound {
		panic("Fatal: Could not estabilsh connection with Tendermint API.")
	}

	// Get IPFS Proxy API
	for _, addr := range conf.ClientConfig().IpfsNodes {
		ipfsAddr := strings.Replace(conf.ClientConfig().IpfsProxyAddr, "$IpfsNode", addr, 1)
		fmt.Println("Trying to connect to (IPFS addr): " + ipfsAddr)

		if response, err := rootAPI.GetBase().IPFSClient.Post(ipfsAddr + conf.ClientConfig().IpfsIsupEndpoint, "application/json", nil); err == nil {
			dat, err := ioutil.ReadAll(response.Body);
			fmt.Printf("Isup: %s\n",dat)
			if err == nil /*&& string(dat) == "dHJ1ZQ=="*/ {
				ipfsProxyFound = true
				rootAPI.GetBase().IPFSAddr = ipfsAddr
				break
			}
		}
	}

	if !ipfsProxyFound {
		panic("Fatal: Could not estabilsh connection with IPFS Proxy API.")
	}

	return rootAPI
}
