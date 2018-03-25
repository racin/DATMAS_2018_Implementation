package main

import (
	"fmt"
	"io/ioutil"
	mh "github.com/multiformats/go-multihash"
	//b58 "github.com/ipfs/go-blocks/Godeps/_workspace/src/github.com/jbenet/go-base58"
	b58 "github.com/mr-tron/base58/base58"
	"errors"
	"runtime/debug"
	"path/filepath"
	"io"
	"time"
	"strings"
	"os"
	"math/rand"
)

// DefaultIpfsHash is the current default hash function used by IPFS.
const DefaultIpfsHash = mh.SHA2_256

// Debug is a global flag for debugging.
var Debug bool

// ErrNotImplemented signifies a function has not been implemented yet.
var ErrNotImplemented = errors.New("Error: not implemented yet.")

// ErrTimeout implies that a timeout has been triggered
var ErrTimeout = errors.New("Error: Call timed out.")

// ErrSearchIncomplete implies that a search type operation didnt
// find the expected node, but did find 'a' node.
var ErrSearchIncomplete = errors.New("Error: Search Incomplete.")

// ErrCast is returned when a cast fails AND the program should not panic.
func ErrCast() error {
	debug.PrintStack()
	return errCast
}

var errCast = errors.New("cast error")

// ExpandPathnames takes a set of paths and turns them into absolute paths
func ExpandPathnames(paths []string) ([]string, error) {
	var out []string
	for _, p := range paths {
		abspath, err := filepath.Abs(p)
		if err != nil {
			return nil, err
		}
		out = append(out, abspath)
	}
	return out, nil
}

type randGen struct {
	rand.Rand
}

// NewTimeSeededRand returns a random bytes reader
// which has been initialized with the current time.
func NewTimeSeededRand() io.Reader {
	src := rand.NewSource(time.Now().UnixNano())
	return &randGen{
		Rand: *rand.New(src),
	}
}

// NewSeededRand returns a random bytes reader
// initialized with the given seed.
func NewSeededRand(seed int64) io.Reader {
	src := rand.NewSource(seed)
	return &randGen{
		Rand: *rand.New(src),
	}
}

func (r *randGen) Read(p []byte) (n int, err error) {
	for i := 0; i < len(p); i++ {
		p[i] = byte(r.Rand.Intn(255))
	}
	return len(p), nil
}

// GetenvBool is the way to check an env var as a boolean
func GetenvBool(name string) bool {
	v := strings.ToLower(os.Getenv(name))
	return v == "true" || v == "t" || v == "1"
}

// MultiErr is a util to return multiple errors
type MultiErr []error

func (m MultiErr) Error() string {
	if len(m) == 0 {
		return "no errors"
	}

	s := "Multiple errors: "
	for i, e := range m {
		if i != 0 {
			s += ", "
		}
		s += e.Error()
	}
	return s
}

// Partition splits a subject 3 parts: prefix, separator, suffix.
// The first occurrence of the separator will be matched.
// ie. Partition("Ready, steady, go!", ", ") -> ["Ready", ", ", "steady, go!"]
func Partition(subject string, sep string) (string, string, string) {
	if i := strings.Index(subject, sep); i != -1 {
		return subject[:i], subject[i : i+len(sep)], subject[i+len(sep):]
	}
	return subject, "", ""
}

// RPartition splits a subject 3 parts: prefix, separator, suffix.
// The last occurrence of the separator will be matched.
// ie. RPartition("Ready, steady, go!", ", ") -> ["Ready, steady", ", ", "go!"]
func RPartition(subject string, sep string) (string, string, string) {
	if i := strings.LastIndex(subject, sep); i != -1 {
		return subject[:i], subject[i : i+len(sep)], subject[i+len(sep):]
	}
	return subject, "", ""
}

// Hash is the global IPFS hash function. uses multihash SHA2_256, 256 bits
func Hash(data []byte) mh.Multihash {
	h, err := mh.Sum(data, DefaultIpfsHash, -1)
	if err != nil {
		// this error can be safely ignored (panic) because multihash only fails
		// from the selection of hash function. If the fn + length are valid, it
		// won't error.
		panic("multihash failed to hash using SHA2_256.")
	}
	return h
}

// IsValidHash checks whether a given hash is valid (b58 decodable, len > 0)
func IsValidHash(s string) bool {
	out, err := b58.Decode(s)
	if err != nil {
		return false
	}
	_, err = mh.Cast(out)
	return err == nil
}

// XOR takes two byte slices, XORs them together, returns the resulting slice.
func XOR(a, b []byte) []byte {
	c := make([]byte, len(a))
	for i := 0; i < len(a); i++ {
		c[i] = a[i] ^ b[i]
	}
	return c
}
func main(){
	/*buf:= []byte("multihash")//hex.DecodeString("0beec7b5ea3f0fdbc95d0dd47f3c5bc275da8a33")
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
	fmt.Println("Wait!")*/
	data, _ := ioutil.ReadFile("uis.sh")
	mHash := Hash([]byte("multihash"))
	fmt.Printf("%s\n", mHash.B58String())
	mHash2 := Hash(data)
	fmt.Printf("%s\n", mHash2.B58String())
	fmt.Printf("File contents: %s", data)
}
/*
// Hash is the global IPFS hash function. uses multihash SHA2_256, 256 bits
func Hash(data []byte) mh.Multihash {
	h, err := mh.Sum(data, mh.SHA2_256, -1)
	if err != nil {
		// this error can be safely ignored (panic) because multihash only fails
		// from the selection of hash function. If the fn + length are valid, it
		// won't error.
		panic("multihash failed to hash using SHA2_256.")
	}
	return h
}*/
/*
// IsValidHash checks whether a given hash is valid (b58 decodable, len > 0)
func IsValidHash(s string) bool {
	//out := b58.Decode(s)
	out, _ := mh.FromB58String(s)
	if out == nil || len(out) == 0 {
		return false
	}
	_, err := mh.Cast(out)
	if err != nil {
		return false
	}
	return true
}*/