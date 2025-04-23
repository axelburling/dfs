/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package master

import (
	"context"
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
	"math/rand"

	"github.com/axelburling/dfs/pkg/master/grpc/client"
	"github.com/axelburling/dfs/pkg/master/grpc/pb"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// registerCmd represents the register command
var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := client.NewClient("localhost:4001")

		if err != nil {
			panic(err)
		}

		id := uuid.New()

		host := make([]byte, 32)

		_, err = crand.Read(host)
		if err != nil {
			panic(err)
		}

		res, err := c.Register(context.Background(), &pb.Node{
			Id: &pb.UUID{
				Value: id[:],
			},
			Addr: "http://localhost:4002/",
			GrpcAddr: "localhost:4004",
			TotalSpace: rand.Int63n(10000000),
			FreeSpace: rand.Int63n(100000),
			Hostname: hex.EncodeToString(host),
			ReadOnly: false,
		})

		if err != nil {
			panic(err)
		}

		fmt.Println(res)
	},
}

func init() {
	MasterCmd.AddCommand(registerCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// registerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// registerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
