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
func getAPI() client.API {
	var api client.API
	var apiOk bool = false
	var remoteAddr string
	fmt.Printf("%+v\n", conf.ClientConfig().TendermintNodes)
	for _, addr := range conf.ClientConfig().TendermintNodes {
		api = client.NewAPI(strings.Replace(conf.ClientConfig().RemoteAddr, "$TmNode", addr, 1))
		fmt.Println("Trying to connect to: " + strings.Replace(conf.ClientConfig().RemoteAddr, "$TmNode", addr, 1))
		if _, err := api.GetBase().TM.Status(); err == nil {
			remoteAddr = addr
			apiOk = true
			break
		}
	}

	conf.ClientConfig().RemoteAddr = strings.Replace(conf.ClientConfig().RemoteAddr, "$TmNode", remoteAddr, 1)
	conf.ClientConfig().UploadAddr = strings.Replace(conf.ClientConfig().UploadAddr, "$TmNode", remoteAddr, 1)
	if !apiOk {
		panic("Fatal: Could not estabilsh connection with API.")
	}

	return api
}
