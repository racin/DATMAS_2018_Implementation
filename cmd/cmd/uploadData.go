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
	"encoding/json"
	"bytes"
)

// getAccountCmd represents the getAccount command
var uploadCmd = &cobra.Command{
	Use:   "upload [file] [name] [description]",
	Short: "upload data",
	Long:  `Upload data.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			log.Fatal("Not enough arguments.")
		}

		// File and Name is required parameters.

		filePath := args[0];
		file, err := os.Open(filePath)
		if err != nil {
			log.Fatal("Could not open file. Error: ", err.Error())
		}

		fileHash, err := crypto.IPFSHashFile(filePath)
		if err != nil {
			log.Fatal("Could not hash file. Error: ", err.Error())
		}

		stranc := getSignedTransaction(app.UploadData, fileHash)
		result, err := getAPI().BeginUploadData(stranc)

		if result != types.CodeType_BCFSBeginUploadOK {
			log.Fatal(err.Error())

		}

		byteArr, _ := json.Marshal(stranc)
		fmt.Println("Data hash registered in application")
		values := map[string]io.Reader{
			"file":    file,
			"transaction": bytes.NewReader(byteArr),
		}

		res := getAPI().EndUploadData(&values)

		if res.Codetype != types.CodeType_OK {
			log.Fatal("Error with upload. ", res.Message)
		}

		// Start timeout to wait for the transaction be put on the ledger.
		fmt.Println("File successfully uploaded.", err)
	},
}

func init() {
	dataCmd.AddCommand(uploadCmd)
}