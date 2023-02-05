/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>

*/
package main

import (
	subcommand "github.com/practice/golang-line/cmd"
	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use: "ef",
		RunE: func(cmd *cobra.Command, args []string) error {
			// if versionF {
			// 	version.PrintVersion()
			// 	return nil
			// }
			return cmd.Usage()
		},
	}
	cmd.AddCommand(subcommand.MonoCmd)

	cmd.Execute()
}
