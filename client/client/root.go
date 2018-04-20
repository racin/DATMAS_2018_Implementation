package client

import (
	"fmt"
	conf "github.com/racin/DATMAS_2018_Implementation/configuration"
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
	"encoding/json"
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

func subToNewBlock(newBlock chan interface{}) error {
	return TheClient.TMClient.Subscribe(context.Background(), "bcfs-client", tmtypes.EventQueryNewBlock, newBlock)
}

func getSignedTransaction(txtype types.TransactionType, data interface{}) (stranc *crypto.SignedStruct) {
	fmt.Printf("%+v\n", data)
	fmt.Printf("%+v\n", TheClient.fingerprint)
	fmt.Printf("%+v\n", txtype)
	fmt.Printf("%+v\n", TheClient.privKey)
	stranc, err := crypto.SignStruct(types.NewTx(data, TheClient.fingerprint, txtype), TheClient.privKey);
	if err != nil {
		panic("Could not sign transaction. Private/Public key pair may not match. Use the --generateKeys to generate a new one. Error: " + err.Error())
	}
	return
}