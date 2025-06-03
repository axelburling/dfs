/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package pubsub

import (
	"fmt"
	"math/rand/v2"

	"github.com/axelburling/dfs/pkg/pubsub/client/msg"
	"github.com/spf13/cobra"
)

// messageCmd represents the message command
var messageCmd = &cobra.Command{
	Use:   "message",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		ms, err := msg.Create(msg.NodeAdd, msg.NodeAddReq{
			ID:          "lkndbkndbn",
			Address:     "http://localhost:4000",
			GrpcAddress: "localhost:4001",
			Hostname:    "computer",
			IsHealthy:   true,
			TotalSpace:  rand.Int64N(100000000),
			FreeSpace:   rand.Int64N(1000000),
			Readonly:    false,
		})

		if err != nil {
			panic(err)
		}

		d, err := msg.Decode[msg.NodeAddReq](ms)

		if err != nil {
			panic(err)
		}

		fmt.Println(d)
	},
}

func init() {
	PubsubCmd.AddCommand(messageCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// messageCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// messageCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
