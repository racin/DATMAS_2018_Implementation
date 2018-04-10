package crypto

import (
	"os"
	"context"
	"io/ioutil"
	"bytes"

	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/core/coreunix"
	"gx/ipfs/QmceUdzxkimdYsgtX733uNgzf1DLHyBKN6ehGSp85ayppM/go-ipfs-cmdkit/files"
	"github.com/ipfs/go-ipfs/repo"
	"github.com/ipfs/go-ipfs/repo/config"
	ds2 "github.com/ipfs/go-ipfs/thirdparty/datastore2"
	mh "github.com/multiformats/go-multihash"
	"fmt"
	"crypto/md5"
	"golang.org/x/crypto/ssh"
)
const HashFunction = mh.SHA2_256

func GetFingerPrint(key *Keys) (string, error){
	if k, err := ssh.NewPublicKey(key.public); err != nil {
		return "", fmt.Errorf("Problems type asserting Public key: %s\n", err.Error())
	} else {
		return fmt.Sprintf("%x", md5.Sum(k.Marshal())), nil
	}
}
func HashData(data []byte) (string, error) {
	var err error
	if mhash, err := mh.Sum(data, HashFunction, -1); err == nil {
		return mhash.B58String(), nil
	}

	return "", err
}
func HashFile(filePath string) (string, error) {
	var err error
	var data []byte
	if _, err = os.Lstat(filePath); err == nil {
		if data, err = ioutil.ReadFile(filePath); err == nil {
			return HashData(data)
		}
	}

	return "", err
}

func IPFSHashData(data []byte) (string, error) {
	buffer := ioutil.NopCloser(bytes.NewBuffer(data))
	return ipfsHash(files.NewReaderFile("a", "a", buffer, nil))
}
func IPFSHashFile(filePath string) (string, error) {
	stat, err := os.Lstat(filePath)
	if err != nil {
		return "", err
	}
	file, _ := files.NewSerialFile(filePath, filePath, false, stat)

	return ipfsHash(file)
}

func ipfsHash(file files.File) (string, error){
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

		err = adder.AddFile(file)
		if err != nil {
			return
		}
	}()

	return (<-out).(*coreunix.AddedObject).Hash, err
}