package main

import (
	"fmt"
	//conf "github.com/racin/DATMAS_2018_Implementation/configuration"
	"github.com/racin/DATMAS_2018_Implementation/client"
	//"github.com/racin/DATMAS_2018_Implementation/app"
	//"github.com/racin/DATMAS_2018_Implementation/crypto"
)

func main() {
	fmt.Println("a")
	a := &client.BaseClient{}
	fmt.Println(a)
	/*cfg, err := conf.LoadClientConfig()
	if err != nil {
		panic("Could not load configuration. Error: " + err.Error())
	}
	fmt.Println("Main client")
	fmt.Printf("%+v\n", cfg.RemoteEndPoint)

	keys,err := crypto.LoadPrivateKey(cfg.PrivateKey)
	if err != nil {
		panic("Could not load private key. Use the --generateKeys option to generate a new one.")
	}

	fp, err := crypto.GetFingerPrint(keys)
	if err != nil {
		panic("Could not load fingerprint of public key. Use the --generateKeys to generate a new one.")
	}

	tranc := app.NewTx("test", fp, app.UploadData)
	stranc, err := tranc.Sign(keys)
	if err != nil {
		panic("Could not sign transaction. Private/Public key pair may not match. Use the --generateKeys to generate a new one.")
	}

	api := getAPI()
	result := api.BeginUploadData(stranc)
	if result != nil{
		panic("Error with result. Error: " + result.Error())
	} else {
		fmt.Println("CheckTx successfully passed.")
	}*/

}
/*
func getAPI() client.API {
	return client.NewAPI(conf.ClientConfig().RemoteEndPoint)
}*/
