package main

import (
	"fmt"
	//"flag"
	"os"

	srv "github.com/racin/DATMAS_2018_Implementation/server"
	"github.com/racin/DATMAS_2018_Implementation/app"

	abci "github.com/tendermint/abci/types"
	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/abci/server"
	"github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/log"

	"github.com/racin/DATMAS_2018_Implementation/configuration"
)

func NewServer(protoAddr, transport string, app abci.Application) (cmn.Service, error) {
	var s cmn.Service
	var err error
	switch transport {
	case "socket":
		s = srv.NewSocketServer(protoAddr, app)
	case "grpc":
		s = srv.NewGRPCServer(protoAddr, abci.NewGRPCApplication(app))
	default:
		err = fmt.Errorf("Unknown server type %s", transport)
	}
	return s, err
}

func main(){
	// Load the app configuration.
	appconf, err := configuration.LoadAppConfig()
	if err != nil {
		panic("Could not get configuration. Error: " + err.Error())
	}

	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))

	// Create the application - in memory or persisted to disk
	app := app.NewApplication()

	// Start the listener
	srv, err := server.NewServer(appconf.ListenAddr, appconf.RpcType, app)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	srv.SetLogger(logger.With("module", "abci-server"))
	if err := srv.Start(); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// Wait forever
	common.TrapSignal(func() {
		// Cleanup
		srv.Stop()
	})
}
