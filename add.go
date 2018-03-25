package main

import (
	//"errors"
	"fmt"
	"io"
	//"os"
	"strings"

	blockservice "github.com/ipfs/go-ipfs/blockservice"

	core "github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/core/coreunix"
	offline "github.com/ipfs/go-ipfs/exchange/offline"
	dag "github.com/ipfs/go-ipfs/merkledag"
	dagtest "github.com/ipfs/go-ipfs/merkledag/test"
	mfs "github.com/ipfs/go-ipfs/mfs"
	ft "github.com/ipfs/go-ipfs/unixfs"

	/*//bstore "github.com/ipfs/go-ipfs-blockstore"
	mh "github.com/ipfs/go-multihash"
	cmdkit "github.com/ipfs/go-ipfs-cmdkit"
	files "github.com/ipfs/go-ipfs-cmdkit/files"
	//pb "github.com/ipfs/pb"
	cmds "github.com/ipfs/go-ipfs-cmds"
	//"log"*/


	/*bstore "gx/ipfs/QmTVDM4LCSUMFNQzbDLL9zQwp8usE6QHymFdh3h8vL9v6b/go-ipfs-blockstore"
	mh "gx/ipfs/QmZyZDi491cCNTLfAhwcaDii2Kg4pwKRkhqQzURGDvY6ua/go-multihash"
	//cmdkit "gx/ipfs/QmceUdzxkimdYsgtX733uNgzf1DLHyBKN6ehGSp85ayppM/go-ipfs-cmdkit"
	cmdkit "github.com/ipfs/go-ipfs-cmdkit"
	files "gx/ipfs/QmceUdzxkimdYsgtX733uNgzf1DLHyBKN6ehGSp85ayppM/go-ipfs-cmdkit/files"
	//pb "gx/ipfs/QmeWjRodbcZFKe5tMN7poEx3izym6osrLSnTLf9UjJZBbs/pb"
	cmds "gx/ipfs/QmfAkMSt9Fwzk48QDJecPcwCUjnf2uG7MLnmCGTp4C6ouL/go-ipfs-cmds"

	pb "github.com/ipfs/go-ipfs/exchange/bitswap/message/pb"
	"gx/ipfs/QmcZfnkapfECQGcLZaf9B79NRg7cRa9EnZh4LSbkCzwNvY/go-cid"*/

	//"github.com/racin/ipfs-cluster/config"
	//"github.com/ipfs/go-ipfs/commands"
	//cid "gx/ipfs/QmcZfnkapfECQGcLZaf9B79NRg7cRa9EnZh4LSbkCzwNvY/go-cid"
	"github.com/ipfs/go-cid"
	repoconf "github.com/ipfs/go-ipfs/repo/config"
	"github.com/ipfs/go-ipfs/commands"
)

// ErrDepthLimitExceeded indicates that the max depth has been exceded.
var ErrDepthLimitExceeded = fmt.Errorf("depth limit exceeded")

const (
	quietOptionName       = "quiet"
	quieterOptionName     = "quieter"
	silentOptionName      = "silent"
	progressOptionName    = "progress"
	trickleOptionName     = "trickle"
	wrapOptionName        = "wrap-with-directory"
	hiddenOptionName      = "hidden"
	onlyHashOptionName    = "only-hash"
	chunkerOptionName     = "chunker"
	pinOptionName         = "pin"
	rawLeavesOptionName   = "raw-leaves"
	noCopyOptionName      = "nocopy"
	fstoreCacheOptionName = "fscache"
	cidVersionOptionName  = "cid-version"
	hashOptionName        = "hash"
)

const adderOutChanSize = 8
func Run (req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) {
		n, err := GetNode(env)
		if err != nil {
			return
		}

		//cfg, err := n.Repo.Config()

		hash := true
		cidVer := 0
		hashFunStr := "sha2-256"
		prefix, err := dag.PrefixForCidVersion(cidVer)
		prefix2 := cid.Prefix(prefix)
		if err != nil {
			return
		}

		hashFunCode, ok := mh.Names[strings.ToLower(hashFunStr)]
		if !ok {
			res.SetError(fmt.Errorf("unrecognized hash function: %s", strings.ToLower(hashFunStr)), cmdkit.ErrNormal)
			return
		}

		prefix2.MhType = hashFunCode
		prefix2.MhLength = -1

		if hash {
			nilnode, err := core.NewNode(n.Context(), &core.BuildCfg{
				//TODO: need this to be true or all files
				// hashed will be stored in memory!
				NilRepo: true,
			})
			if err != nil {
				return
			}
			n = nilnode
		}

		addblockstore := n.Blockstore

		exch := n.Exchange
		local, _ := req.Options["local"].(bool)
		if local {
			exch = offline.Exchange(addblockstore)
		}

		bserv := blockservice.New(addblockstore, exch) // hash security 001
		dserv := dag.NewDAGService(bserv)

		outChan := make(chan interface{}, adderOutChanSize)

		fileAdder, err := coreunix.NewAdder(req.Context, n.Pinning, n.Blockstore, dserv)
		if err != nil {
			return
		}

		fileAdder.Out = outChan
		fileAdder.Chunker = "size-262144"
		fileAdder.Progress = false
		fileAdder.Hidden = true
		fileAdder.Trickle = false
		fileAdder.Wrap = false
		fileAdder.Pin = true
		fileAdder.Silent = false
		fileAdder.RawLeaves = false
		fileAdder.NoCopy = false
		fileAdder.Prefix = &prefix

		if hash {
			md := dagtest.Mock()
			emptyDirNode := ft.EmptyDirNode()
			// Use the same prefix for the "empty" MFS root as for the file adder.
			emptyDirNode.Prefix = *fileAdder.Prefix
			mr, err := mfs.NewRoot(req.Context, md, emptyDirNode, nil)
			if err != nil {
				return
			}

			fileAdder.SetMfsRoot(mr)
		}

		addAllAndPin := func(f files.File) error {
			// Iterate over each top-level file and add individually. Otherwise the
			// single files.File f is treated as a directory, affecting hidden file
			// semantics.
			for {
				file, err := f.NextFile()
				if err == io.EOF {
					// Finished the list of files.
					break
				} else if err != nil {
					return err
				}
				if err := fileAdder.AddFile(file); err != nil {
					return err
				}
			}

			// copy intermediary nodes from editor to our actual dagservice
			_, err := fileAdder.Finalize()
			if err != nil {
				return err
			}

			if hash {
				return nil
			}

			return fileAdder.PinRoot()
		}

		errCh := make(chan error)
		go func() {
			var err error
			defer func() { errCh <- err }()
			defer close(outChan)
			err = addAllAndPin(req.Files)
		}()

		defer res.Close()

		err = res.Emit(outChan)
		if err != nil {
			//log.Error(err)
			return
		}
		err = <-errCh
		if err != nil {
			res.SetError(err, cmdkit.ErrNormal)
		}
		return
	}
	/*
func PostRun()  {

}
	PostRun: cmds.PostRunMap{
		cmds.CLI: func(req *cmds.Request, re cmds.ResponseEmitter) cmds.ResponseEmitter {
			reNext, res := cmds.NewChanResponsePair(req)
			outChan := make(chan interface{})

			sizeChan := make(chan int64, 1)

			sizeFile, ok := req.Files.(files.SizeFile)
			if ok {
				// Could be slow.
				go func() {
					size, err := sizeFile.Size()
					if err != nil {
						// see comment above
						return
					}

					sizeChan <- size
				}()
			} else {
				// we don't need to error, the progress bar just
				// won't know how big the files are
			}

			progressBar := func(wait chan struct{}) {
				defer close(wait)

				quiet, _ := req.Options[quietOptionName].(bool)
				quieter, _ := req.Options[quieterOptionName].(bool)
				quiet = quiet || quieter

				progress, _ := req.Options[progressOptionName].(bool)

				var bar *pb.ProgressBar
				if progress {
					bar = pb.New64(0).SetUnits(pb.U_BYTES)
					bar.ManualUpdate = true
					bar.ShowTimeLeft = false
					bar.ShowPercent = false
					bar.Output = os.Stderr
					bar.Start()
				}

				lastFile := ""
				lastHash := ""
				var totalProgress, prevFiles, lastBytes int64

			LOOP:
				for {
					select {
					case out, ok := <-outChan:
						if !ok {
							if quieter {
								fmt.Fprintln(os.Stdout, lastHash)
							}

							break LOOP
						}
						output := out.(*coreunix.AddedObject)
						if len(output.Hash) > 0 {
							lastHash = output.Hash
							if quieter {
								continue
							}

							if progress {
								// clear progress bar line before we print "added x" output
								fmt.Fprintf(os.Stderr, "\033[2K\r")
							}
							if quiet {
								fmt.Fprintf(os.Stdout, "%s\n", output.Hash)
							} else {
								fmt.Fprintf(os.Stdout, "added %s %s\n", output.Hash, output.Name)
							}

						} else {
							if !progress {
								continue
							}

							if len(lastFile) == 0 {
								lastFile = output.Name
							}
							if output.Name != lastFile || output.Bytes < lastBytes {
								prevFiles += lastBytes
								lastFile = output.Name
							}
							lastBytes = output.Bytes
							delta := prevFiles + lastBytes - totalProgress
							totalProgress = bar.Add64(delta)
						}

						if progress {
							bar.Update()
						}
					case size := <-sizeChan:
						if progress {
							bar.Total = size
							bar.ShowPercent = true
							bar.ShowBar = true
							bar.ShowTimeLeft = true
						}
					case <-req.Context.Done():
						// don't set or print error here, that happens in the goroutine below
						return
					}
				}
			}

			go func() {
				// defer order important! First close outChan, then wait for output to finish, then close re
				defer re.Close()

				if e := res.Error(); e != nil {
					defer close(outChan)
					re.SetError(e.Message, e.Code)
					return
				}

				wait := make(chan struct{})
				go progressBar(wait)

				defer func() { <-wait }()
				defer close(outChan)

				for {
					v, err := res.Next()
					if !cmds.HandleError(err, res, re) {
						break
					}

					select {
					case outChan <- v:
					case <-req.Context.Done():
						re.SetError(req.Context.Err(), cmdkit.ErrNormal)
						return
					}
				}
			}()

			return reNext
		},
	},
	Type: coreunix.AddedObject{},
}*/
// GetNode extracts the node from the environment.
func GetNode(env interface{}) (*core.IpfsNode, error) {
	ctx, ok := env.(*commands.Context)
	if !ok {
		return nil, fmt.Errorf("expected env to be of type %T, got %T", ctx, env)
	}

	return ctx.GetNode()
}

// GetConfig extracts the config from the environment.
func GetConfig(env interface{}) (*repoconf.Config, error) {
	ctx, ok := env.(*commands.Context)
	if !ok {
		return nil, fmt.Errorf("expected env to be of type %T, got %T", ctx, env)
	}

	return ctx.GetConfig()
}
