package client

import (
	"fmt"
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"github.com/spf13/cobra"

	"os"
	"github.com/racin/DATMAS_2018_Implementation/types"
	tmtypes "github.com/tendermint/tendermint/types"
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
	cobra.OnInitialize(NewClient)
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Config file (default is $HOME/.bcfs/clientConfig)")
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// TODO: Fix this.
func subToNewBlock(newBlock chan interface{}) error {
	err := TheClient.TMClient.Start()
	if err != nil {
		fmt.Println("Error starting: " + err.Error())
	}
	return TheClient.TMClient.Subscribe(context.Background(), "bcfs-client", tmtypes.EventQueryNewBlock, newBlock)
}

func getSignedTransaction(txtype types.TransactionType, data interface{}) (stranc *crypto.SignedStruct) {
	fmt.Printf("%+v\n", data)
	fmt.Printf("%+v\n", TheClient.fingerprint)
	fmt.Printf("%+v\n", txtype)
	fmt.Printf("%+v\n", TheClient.privKey)
	tx := types.NewTx(data, TheClient.fingerprint, txtype)
	fmt.Printf("Hash of tranc: %v\n", crypto.HashStruct(tx))
	stranc, err := crypto.SignStruct(tx, TheClient.privKey);
	if err != nil {
		panic("Could not sign transaction. Private/Public key pair may not match. Use the --generateKeys to generate a new one. Error: " + err.Error())
	}
	return
}