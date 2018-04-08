package cmd

import (
	"github.com/spf13/cobra"
	"log"
	"github.com/racin/DATMAS_2018_Implementation/app"
	"fmt"
	"os"
	"github.com/racin/DATMAS_2018_Implementation/crypto"
	"github.com/racin/DATMAS_2018_Implementation/types"
	"io"
	"strings"
)

// getAccountCmd represents the getAccount command
var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "upload data",
	Long:  `Upload data.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			log.Fatal("Not enough arguments.")
		}

		filePath := args[0];
		file, err := openFile(filePath)
		if err != nil {
			fmt.Println("Could not open file. Error: ", err.Error())
		}

		fileHash, err := crypto.IPFSHashFile(filePath)
		if err != nil {
			fmt.Println("Could not hash file. Error: ", err.Error())
		}

		stranc := getSignedTransaction(app.UploadData, fileHash)
		result, err := getAPI().BeginUploadData(stranc)

		if result != types.CodeType_BCFSBeginUploadOK {
			fmt.Println(err.Error())
			return;

		}
		fmt.Println("Data hash registered in application")
		values := map[string]io.Reader{
			"files":    file,
			"dataHash": strings.NewReader(fileHash),
		}
		stranc = getSignedTransaction(app.UploadData, values)
		result, err = getAPI().EndUploadData(stranc)

		if result != types.CodeType_OK {
			fmt.Println("Error with upload. ", err)
			return
		}

		// Start timeout to wait for the transaction be put on the ledger.
		fmt.Println("File successfully uploaded.", err)
	},
}

func openFile(filePath string) (*os.File, error){
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	return file, nil
}
func init() {
	dataCmd.AddCommand(uploadCmd)
}
