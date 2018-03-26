// Package cid implements the Content-IDentifiers specification
// (https://github.com/ipld/cid) in Go. CIDs are
// self-describing content-addressed identifiers useful for
// distributed information systems. CIDs are used in the IPFS
// (https://ipfs.io) project ecosystem.
//
// CIDs have two major versions. A CIDv0 corresponds to a multihash of type
// DagProtobuf, is deprecated and exists for compatibility reasons. Usually,
// CIDv1 should be used.
//
// A CIDv1 has four parts:
//
//     <cidv1> ::= <multibase-prefix><cid-version><multicodec-packed-content-type><multihash-content-address>
//
// As shown above, the CID implementation relies heavily on Multiformats,
// particularly Multibase
// (https://github.com/multiformats/go-multibase), Multicodec
// (https://github.com/multiformats/multicodec) and Multihash
// implementations (https://github.com/multiformats/go-multihash).
package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"context"
	"github.com/ipfs/go-ipfs/repo"
	"github.com/ipfs/go-ipfs/repo/config"
	//"io"
	"io/ioutil"
	//"github.com/ipfs/go-ipfs/pin/gc"

	mbase "github.com/multiformats/go-multibase"
	//mh "github.com/multiformats/go-multihash"
	mh "gx/ipfs/QmZyZDi491cCNTLfAhwcaDii2Kg4pwKRkhqQzURGDvY6ua/go-multihash"
	blocks "gx/ipfs/Qmej7nf81hi2x2tvjRBF3mcp74sQyuDH4VMYDGd1YtXjb2/go-block-format"
	//cid "github.com/ipfs/go-cid"
	cid "gx/ipfs/QmcZfnkapfECQGcLZaf9B79NRg7cRa9EnZh4LSbkCzwNvY/go-cid"
	file "github.com/ipfs/go-ipfs-cmdkit/files"

	"os"
	//"github.com/ipfs/go-ipfs/blockservice"
	//"github.com/ipfs/go-ipfs/commands"
	//core "github.com/ipfs/go-ipfs/core"
	//dag "github.com/ipfs/go-ipfs/merkledag"

	//"time"
	ds2 "github.com/ipfs/go-ipfs/thirdparty/datastore2"
	"github.com/ipfs/go-ipfs/core"
	coreunix "github.com/ipfs/go-ipfs/core/coreunix"
	files "gx/ipfs/QmceUdzxkimdYsgtX733uNgzf1DLHyBKN6ehGSp85ayppM/go-ipfs-cmdkit/files"

/*
	bstore "gx/ipfs/QmTVDM4LCSUMFNQzbDLL9zQwp8usE6QHymFdh3h8vL9v6b/go-ipfs-blockstore"
	blockservice "github.com/ipfs/go-ipfs/blockservice"
	"github.com/ipfs/go-ipfs-blockstore"
	"github.com/ipfs/go-datastore"
	cmds "gx/ipfs/QmfAkMSt9Fwzk48QDJecPcwCUjnf2uG7MLnmCGTp4C6ouL/go-ipfs-cmds"*/
)

// UnsupportedVersionString just holds an error message
const UnsupportedVersionString = "<unsupported cid version>"

var (
	// ErrVarintBuffSmall means that a buffer passed to the cid parser was not
	// long enough, or did not contain an invalid cid
	ErrVarintBuffSmall = errors.New("reading varint: buffer too small")

	// ErrVarintTooBig means that the varint in the given cid was above the
	// limit of 2^64
	ErrVarintTooBig = errors.New("reading varint: varint bigger than 64bits" +
		" and not supported")

	// ErrCidTooShort means that the cid passed to decode was not long
	// enough to be a valid Cid
	ErrCidTooShort = errors.New("cid too short")

	// ErrInvalidEncoding means that selected encoding is not supported
	// by this Cid version
	ErrInvalidEncoding = errors.New("invalid base encoding")
)

// These are multicodec-packed content types. The should match
// the codes described in the authoritative document:
// https://github.com/multiformats/multicodec/blob/master/table.csv
const (
	Raw = 0x55

	DagProtobuf = 0x70
	DagCBOR     = 0x71

	GitRaw = 0x78

	EthBlock           = 0x90
	EthBlockList       = 0x91
	EthTxTrie          = 0x92
	EthTx              = 0x93
	EthTxReceiptTrie   = 0x94
	EthTxReceipt       = 0x95
	EthStateTrie       = 0x96
	EthAccountSnapshot = 0x97
	EthStorageTrie     = 0x98
	BitcoinBlock       = 0xb0
	BitcoinTx          = 0xb1
	ZcashBlock         = 0xc0
	ZcashTx            = 0xc1
)

// Codecs maps the name of a codec to its type
var Codecs = map[string]uint64{
	"v0":                   DagProtobuf,
	"raw":                  Raw,
	"protobuf":             DagProtobuf,
	"cbor":                 DagCBOR,
	"git-raw":              GitRaw,
	"eth-block":            EthBlock,
	"eth-block-list":       EthBlockList,
	"eth-tx-trie":          EthTxTrie,
	"eth-tx":               EthTx,
	"eth-tx-receipt-trie":  EthTxReceiptTrie,
	"eth-tx-receipt":       EthTxReceipt,
	"eth-state-trie":       EthStateTrie,
	"eth-account-snapshot": EthAccountSnapshot,
	"eth-storage-trie":     EthStorageTrie,
	"bitcoin-block":        BitcoinBlock,
	"bitcoin-tx":           BitcoinTx,
	"zcash-block":          ZcashBlock,
	"zcash-tx":             ZcashTx,
}

// CodecToStr maps the numeric codec to its name
var CodecToStr = map[uint64]string{
	Raw:                "raw",
	DagProtobuf:        "protobuf",
	DagCBOR:            "cbor",
	GitRaw:             "git-raw",
	EthBlock:           "eth-block",
	EthBlockList:       "eth-block-list",
	EthTxTrie:          "eth-tx-trie",
	EthTx:              "eth-tx",
	EthTxReceiptTrie:   "eth-tx-receipt-trie",
	EthTxReceipt:       "eth-tx-receipt",
	EthStateTrie:       "eth-state-trie",
	EthAccountSnapshot: "eth-account-snapshot",
	EthStorageTrie:     "eth-storage-trie",
	BitcoinBlock:       "bitcoin-block",
	BitcoinTx:          "bitcoin-tx",
	ZcashBlock:         "zcash-block",
	ZcashTx:            "zcash-tx",
}
type AddedObject struct {
	Name  string
	Hash  string `json:",omitempty"`
	Bytes int64  `json:",omitempty"`
	Size  string `json:",omitempty"`
}
// DefaultIpfsHash is the current default hash function used by IPFS.
const DefaultIpfsHash = mh.SHA2_256

func HashRacin(data []byte) mh.Multihash {
	h, err := mh.Sum(data, DefaultIpfsHash, -1)
	if err != nil {
		// this error can be safely ignored (panic) because multihash only fails
		// from the selection of hash function. If the fn + length are valid, it
		// won't error.
		panic("multihash failed to hash using SHA2_256.")
	}
	return h
}

func main(){
	filePath := "uis.sh"
	fmt.Println("Racin");

	stat, _ := os.Lstat(filePath)
	f, _ := file.NewSerialFile(filePath,filePath,false, stat)
	data, _ := ioutil.ReadAll(f)
	mHash := HashRacin([]byte("multihash"))
	fmt.Printf("%s\n", mHash.B58String())
	mHash2 := HashRacin(data)
	fmt.Printf("%s\n", mHash2.B58String())
	fmt.Printf("File contents: %s", data)

	p := NewPrefixV1( 	0x01A5, mh.SHA2_256)
	s, _ := p.Sum(data)
	c := cid.NewCidV0( s.hash)
	fmt.Printf("SHash: %s\n", s.hash.B58String())
	pc,_ := Cast(s.hash)
	fmt.Printf("PC: %s\n", pc.String())
	b, _ := blocks.NewBlockWithCid(data, c)
	fmt.Printf("Cid: %s\n", b.RawData())
	fmt.Printf("Prefix1: %s\n", b.Cid().Prefix())

	b281, _ := cid.Parse(b.Cid().String())
	fmt.Printf("Prefix2: %s\n", b281.Hash().B58String())
	json1, _ := b281.MarshalJSON()
	fmt.Printf("B11: %s\n", json1)
	mhd1, _ := mh.Decode([]byte(b.Cid().Hash()))
	fmt.Printf("Decoded: %+v\n", mhd1)

	b28Hash := "QmX2aMQtxnvmm5T3DFQTzSBXLSUfbNsysKnh2SEUjBTt4X"
	b2,_ := cid.Parse(b28Hash)
	mhd, _ := mh.Decode([]byte(b2.Hash()))
	fmt.Printf("Decoded: %+v\n", mhd)
	json, _ := b2.MarshalJSON()
	fmt.Printf("B2: %s\n", json)
	p2 := NewPrefixV0(mh.SHA2_256)
	s2, _ := p2.Sum(data)
	fmt.Printf("%s\n", s2.hash.B58String())

	HashFiles(filePath)
/*
	c := NewCidV1( 	0x01A5, s.hash)
	sc, _ :=
	fmt.Printf("%s\n", s.hash.B58String())
	c2 := NewPrefixV0(mh.SHA2_256)
	sc2, _ := p2.Sum(mHash2)
	fmt.Printf("%s\n", s2.hash.B58String())*/
	//bserv := blockservice.New(addblockstore, exch) // hash security 001
	/*env := cmds.Environment
	addblockstore := bstore.NewGCBlockstore(blockstore.NewBlockstore( datastore.Batching(data)))
	bserv := blockservice.New()
	//a := dag.NewRawNode(data)
	//fmt.Printf("%s\n", a.Cid())
	dserv := dag.NewDAGService(bserv)
	y := make(chan <-)
	z :=*/
}
type error interface {
	Error() string
}
const testPeerID = "QmTFauExutTsy4XP6JbMFcw2Wa9645HJt2bTqL6qYDCKfe"
func HashFiles(filePath string) error{
	r := &repo.Mock{
		C: config.Config{
			Identity: config.Identity{
				PeerID: testPeerID, // required by offline node
			},
		},
		D: ds2.ThreadSafeCloserMapDatastore(),
	}
	node, err := core.NewNode(context.Background(), &core.BuildCfg{Repo: r})
	if err != nil {
		return err
	}

	out := make(chan interface{})
	adder, err := coreunix.NewAdder(context.Background(), node.Pinning, node.Blockstore, node.DAG)
	if err != nil {
		return err
	}
	adder.Out = out

	//dataa := ioutil.NopCloser(bytes.NewBufferString("testfileA"))
	//files.NewSerialFile()
	stat, _ := os.Lstat(filePath)
	rfa, _ := files.NewSerialFile(filePath,filePath,false, stat)

/*	// make two files with pipes so we can 'pause' the add for timing of the test
	piper, pipew := io.Pipe()
	hangfile := files.NewReaderFile("b", "b", piper, nil)

	datad := ioutil.NopCloser(bytes.NewBufferString("testfileD"))
	rfd := files.NewReaderFile("d", "d", datad, nil)

	slf := files.NewSliceFile("files", "files", []files.File{rfa, hangfile, rfd})*/

	addDone := make(chan struct{})
	go func() {
		defer close(addDone)
		defer close(out)
		err := adder.AddFile(rfa)

		if err != nil {
			return
		}

	}()
	fmt.Println("Racin")
	select {
	case o := <-out:
		fmt.Println("Got hash: ", o.(*coreunix.AddedObject).Hash)
	case <-addDone:
		return ErrVarintTooBig
	}
	/*addedHashes := make(map[string]struct{})
	select {
	case o := <-out:
		addedHashes[o.(*coreunix.AddedObject).Hash] = struct{}{}
	case <-addDone:
		return ErrVarintTooBig
	}
	fmt.Printf("Addedhashes: %s\n", addedHashes)*/
	fmt.Println("Racin 2")
	/*var gcout <-chan gc.Result
	gcstarted := make(chan struct{})
	go func() {
		defer close(gcstarted)
		gcout = gc.GC(context.Background(), node.Blockstore, node.Repo.Datastore(), node.Pinning, nil)
	}()

	// gc shouldnt start until we let the add finish its current file.
	//pipew.Write([]byte("some data for file b"))

	select {
	case <-gcstarted:
		return ErrCidTooShort
	default:
	}

	time.Sleep(time.Millisecond * 100) // make sure gc gets to requesting lock

	// finish write and unblock gc
	//pipew.Close()

	// receive next object from adder
	o := <-out
	addedHashes[o.(*coreunix.AddedObject).Hash] = struct{}{}

	<-gcstarted

	for r := range gcout {
		if r.Error != nil {
			return err
		}
		if _, ok := addedHashes[r.KeyRemoved.String()]; ok {
			return errors.New("gc'ed a hash we just added")
		}
	}

	var last *cid.Cid
	for a := range out {
		// wait for it to finish
		c, err := cid.Decode(a.(*coreunix.AddedObject).Hash)
		if err != nil {
			return err
		}
		last = c
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	set := cid.NewSet()
	err = dag.EnumerateChildren(ctx, dag.GetLinksWithDAG(node.DAG), last, set.Visit)
	if err != nil {
		return err
	}
*/
	return nil
}/*
func GetNode2(env interface{}) (*core.IpfsNode, error) {
	ctx, ok := env.(*commands.Context)
	if !ok {
		return nil, fmt.Errorf("expected env to be of type %T, got %T", ctx, env)
	}

	return ctx.GetNode()
}*/
// NewCidV0 returns a Cid-wrapped multihash.
// They exist to allow IPFS to work with Cids while keeping
// compatibility with the plain-multihash format used used in IPFS.
// NewCidV1 should be used preferentially.
func NewCidV0(mhash mh.Multihash) *Cid {
	return &Cid{
		version: 0,
		codec:   DagProtobuf,
		hash:    mhash,
	}
}

// NewCidV1 returns a new Cid using the given multicodec-packed
// content type.
func NewCidV1(codecType uint64, mhash mh.Multihash) *Cid {
	return &Cid{
		version: 1,
		codec:   codecType,
		hash:    mhash,
	}
}

// NewPrefixV0 returns a CIDv0 prefix with the specified multihash type.
func NewPrefixV0(mhType uint64) Prefix {
	return Prefix{
		MhType:   mhType,
		MhLength: mh.DefaultLengths[mhType],
		Version:  0,
		Codec:    DagProtobuf,
	}
}

// NewPrefixV1 returns a CIDv1 prefix with the specified codec and multihash
// type.
func NewPrefixV1(codecType uint64, mhType uint64) Prefix {
	return Prefix{
		MhType:   mhType,
		MhLength: mh.DefaultLengths[mhType],
		Version:  1,
		Codec:    codecType,
	}
}

// Cid represents a self-describing content adressed
// identifier. It is formed by a Version, a Codec (which indicates
// a multicodec-packed content type) and a Multihash.
type Cid struct {
	version uint64
	codec   uint64
	hash    mh.Multihash
}

// Parse is a short-hand function to perform Decode, Cast etc... on
// a generic interface{} type.
func Parse(v interface{}) (*Cid, error) {
	switch v2 := v.(type) {
	case string:
		if strings.Contains(v2, "/ipfs/") {
			return Decode(strings.Split(v2, "/ipfs/")[1])
		}
		return Decode(v2)
	case []byte:
		return Cast(v2)
	case mh.Multihash:
		return NewCidV0(v2), nil
	case *Cid:
		return v2, nil
	default:
		return nil, fmt.Errorf("can't parse %+v as Cid", v2)
	}
}

// Decode parses a Cid-encoded string and returns a Cid object.
// For CidV1, a Cid-encoded string is primarily a multibase string:
//
//     <multibase-type-code><base-encoded-string>
//
// The base-encoded string represents a:
//
// <version><codec-type><multihash>
//
// Decode will also detect and parse CidV0 strings. Strings
// starting with "Qm" are considered CidV0 and treated directly
// as B58-encoded multihashes.
func Decode(v string) (*Cid, error) {
	if len(v) < 2 {
		return nil, ErrCidTooShort
	}

	if len(v) == 46 && v[:2] == "Qm" {
		hash, err := mh.FromB58String(v)
		if err != nil {
			return nil, err
		}

		return NewCidV0(hash), nil
	}

	_, data, err := mbase.Decode(v)
	if err != nil {
		return nil, err
	}

	return Cast(data)
}

func uvError(read int) error {
	switch {
	case read == 0:
		return ErrVarintBuffSmall
	case read < 0:
		return ErrVarintTooBig
	default:
		return nil
	}
}

// Cast takes a Cid data slice, parses it and returns a Cid.
// For CidV1, the data buffer is in the form:
//
//     <version><codec-type><multihash>
//
// CidV0 are also supported. In particular, data buffers starting
// with length 34 bytes, which starts with bytes [18,32...] are considered
// binary multihashes.
//
// Please use decode when parsing a regular Cid string, as Cast does not
// expect multibase-encoded data. Cast accepts the output of Cid.Bytes().
func Cast(data []byte) (*Cid, error) {
	if len(data) == 34 && data[0] == 18 && data[1] == 32 {
		h, err := mh.Cast(data)
		if err != nil {
			return nil, err
		}

		return &Cid{
			codec:   DagProtobuf,
			version: 0,
			hash:    h,
		}, nil
	}

	vers, n := binary.Uvarint(data)
	if err := uvError(n); err != nil {
		return nil, err
	}

	if vers != 0 && vers != 1 {
		return nil, fmt.Errorf("invalid cid version number: %d", vers)
	}

	codec, cn := binary.Uvarint(data[n:])
	if err := uvError(cn); err != nil {
		return nil, err
	}

	rest := data[n+cn:]
	h, err := mh.Cast(rest)
	if err != nil {
		return nil, err
	}

	return &Cid{
		version: vers,
		codec:   codec,
		hash:    h,
	}, nil
}

// Type returns the multicodec-packed content type of a Cid.
func (c *Cid) Type() uint64 {
	return c.codec
}

// String returns the default string representation of a
// Cid. Currently, Base58 is used as the encoding for the
// multibase string.
func (c *Cid) String() string {
	switch c.version {
	case 0:
		return c.hash.B58String()
	case 1:
		mbstr, err := mbase.Encode(mbase.Base58BTC, c.bytesV1())
		if err != nil {
			panic("should not error with hardcoded mbase: " + err.Error())
		}

		return mbstr
	default:
		panic("not possible to reach this point")
	}
}

// String returns the string representation of a Cid
// encoded is selected base
func (c *Cid) StringOfBase(base mbase.Encoding) (string, error) {
	switch c.version {
	case 0:
		if base != mbase.Base58BTC {
			return "", ErrInvalidEncoding
		}
		return c.hash.B58String(), nil
	case 1:
		return mbase.Encode(base, c.bytesV1())
	default:
		panic("not possible to reach this point")
	}
}

// Hash returns the multihash contained by a Cid.
func (c *Cid) Hash() mh.Multihash {
	return c.hash
}

// Bytes returns the byte representation of a Cid.
// The output of bytes can be parsed back into a Cid
// with Cast().
func (c *Cid) Bytes() []byte {
	switch c.version {
	case 0:
		return c.bytesV0()
	case 1:
		return c.bytesV1()
	default:
		panic("not possible to reach this point")
	}
}

func (c *Cid) bytesV0() []byte {
	return []byte(c.hash)
}

func (c *Cid) bytesV1() []byte {
	// two 8 bytes (max) numbers plus hash
	buf := make([]byte, 2*binary.MaxVarintLen64+len(c.hash))
	n := binary.PutUvarint(buf, c.version)
	n += binary.PutUvarint(buf[n:], c.codec)
	cn := copy(buf[n:], c.hash)
	if cn != len(c.hash) {
		panic("copy hash length is inconsistent")
	}

	return buf[:n+len(c.hash)]
}

// Equals checks that two Cids are the same.
// In order for two Cids to be considered equal, the
// Version, the Codec and the Multihash must match.
func (c *Cid) Equals(o *Cid) bool {
	return c.codec == o.codec &&
		c.version == o.version &&
		bytes.Equal(c.hash, o.hash)
}

// UnmarshalJSON parses the JSON representation of a Cid.
func (c *Cid) UnmarshalJSON(b []byte) error {
	if len(b) < 2 {
		return fmt.Errorf("invalid cid json blob")
	}
	obj := struct {
		CidTarget string `json:"/"`
	}{}
	err := json.Unmarshal(b, &obj)
	if err != nil {
		return err
	}

	if obj.CidTarget == "" {
		return fmt.Errorf("cid was incorrectly formatted")
	}

	out, err := Decode(obj.CidTarget)
	if err != nil {
		return err
	}

	c.version = out.version
	c.hash = out.hash
	c.codec = out.codec
	return nil
}

// MarshalJSON procudes a JSON representation of a Cid, which looks as follows:
//
//    { "/": "<cid-string>" }
//
// Note that this formatting comes from the IPLD specification
// (https://github.com/ipld/specs/tree/master/ipld)
func (c *Cid) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("{\"/\":\"%s\"}", c.String())), nil
}

// KeyString casts the result of cid.Bytes() as a string, and returns it.
func (c *Cid) KeyString() string {
	return string(c.Bytes())
}

// Loggable returns a Loggable (as defined by
// https://godoc.org/github.com/ipfs/go-log).
func (c *Cid) Loggable() map[string]interface{} {
	return map[string]interface{}{
		"cid": c,
	}
}

// Prefix builds and returns a Prefix out of a Cid.
func (c *Cid) Prefix() Prefix {
	dec, _ := mh.Decode(c.hash) // assuming we got a valid multiaddr, this will not error
	return Prefix{
		MhType:   dec.Code,
		MhLength: dec.Length,
		Version:  c.version,
		Codec:    c.codec,
	}
}

// Prefix represents all the metadata of a Cid,
// that is, the Version, the Codec, the Multihash type
// and the Multihash length. It does not contains
// any actual content information.
type Prefix struct {
	Version  uint64
	Codec    uint64
	MhType   uint64
	MhLength int
}

// Sum uses the information in a prefix to perform a multihash.Sum()
// and return a newly constructed Cid with the resulting multihash.
func (p Prefix) Sum(data []byte) (*Cid, error) {
	hash, err := mh.Sum(data, p.MhType, p.MhLength)
	if err != nil {
		return nil, err
	}

	switch p.Version {
	case 0:
		return NewCidV0(hash), nil
	case 1:
		return NewCidV1(p.Codec, hash), nil
	default:
		return nil, fmt.Errorf("invalid cid version")
	}
}

// Bytes returns a byte representation of a Prefix. It looks like:
//
//     <version><codec><mh-type><mh-length>
func (p Prefix) Bytes() []byte {
	buf := make([]byte, 4*binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, p.Version)
	n += binary.PutUvarint(buf[n:], p.Codec)
	n += binary.PutUvarint(buf[n:], uint64(p.MhType))
	n += binary.PutUvarint(buf[n:], uint64(p.MhLength))
	return buf[:n]
}

// PrefixFromBytes parses a Prefix-byte representation onto a
// Prefix.
func PrefixFromBytes(buf []byte) (Prefix, error) {
	r := bytes.NewReader(buf)
	vers, err := binary.ReadUvarint(r)
	if err != nil {
		return Prefix{}, err
	}

	codec, err := binary.ReadUvarint(r)
	if err != nil {
		return Prefix{}, err
	}

	mhtype, err := binary.ReadUvarint(r)
	if err != nil {
		return Prefix{}, err
	}

	mhlen, err := binary.ReadUvarint(r)
	if err != nil {
		return Prefix{}, err
	}

	return Prefix{
		Version:  vers,
		Codec:    codec,
		MhType:   mhtype,
		MhLength: int(mhlen),
	}, nil
}
