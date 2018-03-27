package crypto

import (
	"os"
	"context"

	"github.com/ipfs/go-ipfs/core"
	coreunix "github.com/ipfs/go-ipfs/core/coreunix"
	files "gx/ipfs/QmceUdzxkimdYsgtX733uNgzf1DLHyBKN6ehGSp85ayppM/go-ipfs-cmdkit/files"
	"github.com/ipfs/go-ipfs/repo"
	"github.com/ipfs/go-ipfs/repo/config"
	ds2 "github.com/ipfs/go-ipfs/thirdparty/datastore2"
	"io/ioutil"
	"bytes"
)

func IPFSHashData(data []byte) (string, error){
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

	go func() {
		defer close(out)
		buffer := ioutil.NopCloser(bytes.NewBuffer(data))

		fileBuf := files.NewReaderFile("a", "a", buffer, nil)

		err = adder.AddFile(fileBuf)
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
