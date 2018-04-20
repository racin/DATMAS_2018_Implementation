package ipfsproxy

import (
	"os"
	"fmt"
)

// Currently not in use. Use Addfilenopin instead
func (proxy *Proxy) IPFSAddFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}

	result, err := proxy.client.IPFS().Add(file)
	fmt.Println(result)
	return err
}
