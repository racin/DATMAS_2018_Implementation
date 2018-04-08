package cmd

import (
	"github.com/spf13/cobra"
	"log"
	"github.com/racin/DATMAS_2018_Implementation/app"
	"fmt"
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

		stranc := getSignedTransaction(app.UploadData,"Racin test")
		result := getAPI().BeginUploadData(stranc)
		if result != nil {
			fmt.Println("Error with result. Error: " + result.Error())
		} else {
			fmt.Println("CheckTx successfully passed.")
		}
	},
}

func init() {
	dataCmd.AddCommand(uploadCmd)
}
