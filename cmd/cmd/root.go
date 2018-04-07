package cmd

import (
	"fmt"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	"github.com/racin/DATMAS_2018_Implementation/client"
	"github.com/racin/DATMAS_2018_Implementation/app"
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"github.com/spf13/cobra"

	"os"
)

var rootCmd = &cobra.Command{
	Use:   "bcfs",
	Short: "Block Chain File System",
	Long: `Implementation of Block Chain File System for 
			Master Thesis in Computer Science at UiS 2018.
			Written by Racin Nygaard.	`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

var cfgFile string
func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Config file (default is $HOME/.bcfs/clientConfig)")
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
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	fmt.Println("a")
	fmt.Println("Main client")

	stranc := getSignedTransaction(app.UploadData,"Racin test")
	result := getAPI().BeginUploadData(stranc)
	if result != nil {
		panic("Error with result. Error: " + result.Error())
	} else {
		fmt.Println("CheckTx successfully passed.")
	}

}

func getSignedTransaction(txtype app.TransactionType, data interface{}) (stranc *app.SignedTransaction) {
	keys, err := crypto.LoadPrivateKey(conf.AppConfig().BasePath + conf.AppConfig().PrivateKey)
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
	var tm *rpcClient.HTTP
	for _, addr := range conf.ClientConfig().TendermintNodes {

		tm := rpcClient.NewHTTP(addr, conf.ClientConfig().WebsocketEndPoint)
		if tm.IsRunning() {
			break;
		}
	}
	return client.NewAPI(conf.ClientConfig().RemoteEndPoint)
}
