/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package master

import (
	logger "github.com/axelburling/dfs/internal/log"
	"github.com/axelburling/dfs/pkg/master"
	"github.com/spf13/cobra"
)

// masterCmd represents the master command
var MasterCmd = &cobra.Command{
	Use:   "master",
	Short: "Start a master instance",
	Long: `The master node is responsible for managing metadata, coordinating nodes,
and handling client requests in the distributed file system.

Responsibilities:
- Tracks all nodes and their available storage capacity.
- Manages file metadata, chunk locations, and replication.
- Ensures high availability and fault tolerance.
- Handles client requests for file uploads, downloads, and metadata retrieval.
- Balances workload across storage nodes.

This command starts the master instance and listens for incoming requests from nodes and clients.`,
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

		m := master.New(log)

		m.Start()
	},
}

func init() {
	MasterCmd.Flags().BoolP("dev", "d", false, "toggle master to development")
}
