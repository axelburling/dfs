/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package crypto

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/axelburling/dfs/internal/log"
	"github.com/axelburling/dfs/pkg/node/storage/crypto"
	"github.com/spf13/cobra"
)

// encryptCmd represents the encrypt command
var encryptCmd = &cobra.Command{
	Use:   "encrypt",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.ParseFlags(args)

		if err != nil {
			panic(err)
		}

		key, err := cmd.Flags().GetString("key")

		if err != nil {
			panic(err)
		}

		path, err := cmd.Flags().GetString("file")

		if err != nil {
			panic(err)
		}

		oPath, err := cmd.Flags().GetString("output")

		if err != nil {
			panic(err)
		}



		k, err := base64.StdEncoding.DecodeString(strings.Trim(key, " "))

		if err != nil {
			panic(err)
		}


		cr, err := crypto.NewCrypto([]byte(k), log.NewLogger(log.Development))

		if err != nil {
			panic(err)
		}

		file, err := os.Open(path)

		val, err := cr.EncryptReader(file)

		if err != nil {
			panic(err)
		}

		f, err := os.Create(oPath)

		if err != nil {
			panic(err)
		}

		n, err := io.Copy(f, val)

		if err != nil {
			panic(err)
		}

		fmt.Printf("written: %v bytes\n", n )


	},
}

func init() {
	CryptoCmd.AddCommand(encryptCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// encryptCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// encryptCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	encryptCmd.Flags().StringP("key", "k", "", "--key <32 bit long key>")

	encryptCmd.Flags().StringP("file", "f", "", "--file <path to file to encrypt>")
	encryptCmd.Flags().StringP("output", "o", "", "--output <path to file to encrypt>")
}
