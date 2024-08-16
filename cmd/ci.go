/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ciCmd represents the ci command
var ciCmd = &cobra.Command{
	Use:   "ci",
	Short: "For CI integrations",
	Long: `
	
	Includes features for better CI integrations such as failure when updates available 
	for pipelines, allowing users to automatically create PRs when updates are present(custom thresholds) and so on.

Not all those who don't update dependencies are lost.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("ci called")
	},
}

func init() {
	checkForUpdatesCmd.AddCommand(ciCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// ciCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// ciCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
