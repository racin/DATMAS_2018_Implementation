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

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	fmt.Println("a")
	cfg, err := conf.LoadClientConfig()
	if err != nil {
		panic("Could not load configuration. Error: " + err.Error())
	}
	fmt.Println("Main client")
	fmt.Printf("%+v\n", cfg.RemoteEndPoint)

	keys, err := crypto.LoadPrivateKey(cfg.BasePath + cfg.PrivateKey)
	if err != nil {
		panic("Could not load private key. Use the --generateKeys option to generate a new one. Error: " + err.Error())
	}

	fp, err := crypto.GetFingerPrint(keys)
	if err != nil {
		panic("Could not load fingerprint of public key. Use the --generateKeys to generate a new one. Error: " + err.Error())
	}

	tranc := app.NewTx("test", fp, app.UploadData)
	stranc, err := tranc.Sign(keys)
	if err != nil {
		panic("Could not sign transaction. Private/Public key pair may not match. Use the --generateKeys to generate a new one. Error: " + err.Error())
	}

	api := getAPI()
	result := api.BeginUploadData(stranc)
	if result != nil {
		panic("Error with result. Error: " + result.Error())
	} else {
		fmt.Println("CheckTx successfully passed.")
	}

}

func getAPI() client.API {
	return client.NewAPI(conf.ClientConfig().RemoteEndPoint)
}
