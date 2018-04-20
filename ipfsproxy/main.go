package main

import (
	"fmt"
	"github.com/racin/DATMAS_2018_Implementation/ipfsproxy/ipfsproxy"
)


func main() {
	proxy := ipfsproxy.NewProxy()
	fmt.Println("Starting IPFS proxy API")
	proxy.StartHTTPAPI()
}