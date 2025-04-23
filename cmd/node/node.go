/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package node

import (
	cmdCrypto "github.com/axelburling/dfs/cmd/node/crypto"
	"github.com/axelburling/dfs/internal/config"
	logger "github.com/axelburling/dfs/internal/log"
	"github.com/axelburling/dfs/pkg/node"
	"github.com/axelburling/dfs/pkg/node/storage"
	"github.com/axelburling/dfs/pkg/node/storage/crypto"
	"github.com/spf13/cobra"
)

// nodeCmd represents the node command
var NodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Start a node instance",
	Long: `A storage node is responsible for storing file chunks and responding
to master and client requests for file storage and retrieval.

Responsibilities:
- Stores and serves file chunks efficiently.
- Reports health status and available storage to the master.
- Replicates chunks as directed by the master.
- Handles data integrity checks and self-healing in case of failures.

This command starts a storage node and connects it to the master instance.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.ParseFlags(args)

		if err != nil {
			panic(err)
		}

		dev, err := cmd.Flags().GetBool("dev")
		if err != nil {
			panic(err)
		}
		var log *logger.Logger

		if dev {
			log = logger.NewLogger(logger.Development)
		} else {
			log = logger.NewLogger(logger.Production)
		}

		dir, err := cmd.Flags().GetString("dir")

		if err != nil {
			panic(err)
		}

		conf := config.New[config.NodeConfig](config.Node, log)

		crypt, err := crypto.NewCrypto(conf.EncryptionKey, log)

		store, err := storage.NewFileStorage(dir, log, crypt)
		if err != nil {
			panic(err)
		}

		no := node.New(conf, "./data", log, store)
		no.Start()
	},
}

func init() {
	NodeCmd.AddCommand(cmdCrypto.CryptoCmd)

	NodeCmd.Flags().String("dir", "data", "where to store the chunks")
	NodeCmd.Flags().BoolP("dev", "d", false, "toggle node to development mode")
}
