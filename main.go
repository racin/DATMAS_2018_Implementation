/*
Package server is used to start a new ABCI server.

It contains two server implementation:
 * gRPC server
 * socket server

*/

package main

import (
	"context"
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
	"github.com/ipfs/go-ipfs/repo"
	"github.com/ipfs/go-ipfs/repo/config"
	ds2 "github.com/ipfs/go-ipfs/thirdparty/datastore2"

	"github.com/ipfs/go-ipfs/core"
	coreunix "github.com/ipfs/go-ipfs/core/coreunix"
	files "gx/ipfs/QmceUdzxkimdYsgtX733uNgzf1DLHyBKN6ehGSp85ayppM/go-ipfs-cmdkit/files"
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
	hash, _ := IPFSHashFile("uis.sh")
	fmt.Println("Hash: " + hash)
	hash2, _ := IPFSHashFile("hash.go")
	fmt.Println("Hash: " + hash2)

	// Wait forever
	/*common.TrapSignal(func() {
		// Cleanup
		srv.Stop()
	})*/
}

func IPFSHashFile(filePath string) (string, error){
	var hash string
	r := &repo.Mock{
		C: config.Config{
			Identity: config.Identity{
				PeerID: "QmTFauExutTsy4XP6JbMFcw2Wa9645HJt2bTqL6qYDCKfe", // required by offline node
			},
		},
		D: ds2.ThreadSafeCloserMapDatastore(),
	}
	node, err := core.NewNode(context.Background(), &core.BuildCfg{Repo: r})
	if err != nil {
		return hash, err
	}

	adder, err := coreunix.NewAdder(context.Background(), node.Pinning, node.Blockstore, node.DAG)
	if err != nil {
		return hash, err
	}
	out := make(chan interface{})
	adder.Out = out

	stat, err := os.Lstat(filePath)
	if err != nil {
		return hash, err
	}


	go func() {
		defer close(out)
		file, _ := files.NewSerialFile(filePath,filePath,false, stat)

		err = adder.AddFile(file)
		if err != nil {
			return
		}
	}()

	select {
		case o := <-out:
			hash = o.(*coreunix.AddedObject).Hash
	}

	return hash, err
}