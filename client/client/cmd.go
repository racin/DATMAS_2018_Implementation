package client

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var RootCmd = &cobra.Command{
	Use:   "bcfs",
	Short: "Block Chain File System",
	Long: `Implementation of Block Chain File System for Master Thesis in Computer Science at UiS 2018.
Written by Racin Nygaard.	`,
}


var cfgFile string
func init() {
	cobra.OnInitialize(NewClient)
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Config file (default is $HOME/.bcfs/clientConfig)")
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}