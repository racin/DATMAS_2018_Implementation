/*
Package server is used to start a new ABCI server.

It contains two server implementation:
 * gRPC server
 * socket server

*/

package main

import (
	"fmt"
	"flag"
	"os"
	srv "github.com/racin/DATMAS_2018_Implementation/server"
	"github.com/racin/DATMAS_2018_Implementation/abci-app"
	"github.com/tendermint/abci/types"
	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/abci/server"
	//"github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/log"
	"github.com/multiformats/go-multihash"
	"encoding/hex"
	"crypto/sha1"
)

func NewServer(protoAddr, transport string, app types.Application) (cmn.Service, error) {
	var s cmn.Service
	var err error
	switch transport {
	case "socket":
		s = srv.NewSocketServer(protoAddr, app)
	case "grpc":
		s = srv.NewGRPCServer(protoAddr, types.NewGRPCApplication(app))
	default:
		err = fmt.Errorf("Unknown server type %s", transport)
	}
	return s, err
}

func main(){
	addrPtr := flag.String("addr", "tcp://0.0.0.0:46658", "Listen address")
	abciPtr := flag.String("abci", "socket", "socket | grpc")
	//storePtr := flag.String("store", "app.ldb", "store path")
	flag.Parse()

	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))

	// Create the application - in memory or persisted to disk
	app := app.NewApplication()

	// Start the listener
	srv, err := server.NewServer(*addrPtr, *abciPtr, app)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	srv.SetLogger(logger.With("module", "abci-server"))
	if err := srv.Start(); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	fmt.Println("Racin har started en app! Transport: " + *abciPtr);
	fmt.Println("Info om app: " + app.Info(types.RequestInfo{Version: "123"}).Data)
/*	buf, _ := hex.DecodeString("48656c6c6f20476f7068657221")
	fmt.Printf("%s\n", buf)
	mHashBuf, _ := multihash.Encode([]byte("multihash"), multihash.SHA2_256)
	mh, _ := multihash.Cast(mHashBuf);
	fmt.Printf("hex: %s\n", hex.EncodeToString(mHashBuf))
	fmt.Println(mh.B58String())
	mHash, _ := multihash.Decode(mHashBuf)
	sha256hex := hex.EncodeToString(mHash.Digest)
	fmt.Printf("obj: %v 0x%x %d %s\n", mHash.Name, mHash.Code, mHash.Length, sha256hex)
	fmt.Println()*/
	// ignores errors for simplicity.
	// don't do that at home.
	// Decode a SHA1 hash to a binary buffer
	//buf, _ := hex.DecodeString("0beec7b5ea3f0fdbc95d0dd47f3c5bc275da8a33")
	buf:= []byte("multihash")//hex.DecodeString("0beec7b5ea3f0fdbc95d0dd47f3c5bc275da8a33")
	// Create a new multihash with it.
	fmt.Printf("%x\n", sha1.Sum(buf))
	z := []byte(sha1.Sum(buf))
	fmt.Println(hex.EncodeToString())
	mHashBuf, _ := multihash.EncodeName(buf, "sha1")
	// Print the multihash as hex string
	fmt.Printf("hex: %s\n", hex.EncodeToString(mHashBuf))

	// Parse the binary multihash to a DecodedMultihash
	mHash, _ := multihash.Decode(mHashBuf)
	// Convert the sha1 value to hex string
	sha1hex := hex.EncodeToString(mHash.Digest)
	// Print all the information in the multihash
	fmt.Printf("obj: %v 0x%x %d %s\n", mHash.Name, mHash.Code, mHash.Length, sha1hex)
	fmt.Println("Wait!")
	// Wait forever
	/*common.TrapSignal(func() {
		// Cleanup
		srv.Stop()
	})*/
}