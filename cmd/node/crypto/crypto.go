/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/spf13/cobra"
)

// cryptoCmd represents the crypto command
var CryptoCmd = &cobra.Command{
	Use:   "crypto",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		key := make([]byte, 32)

		_, err := rand.Read(key)
		if err != nil {
			panic(err)
		}

		fmt.Printf("random generated aes encryption key: %s\n", base64.StdEncoding.EncodeToString(key))
	},
}

func init() {

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cryptoCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cryptoCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
