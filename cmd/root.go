/*
Copyright © 2025 NAME HERE axel.burling@gmail.com
*/
package cmd

import (
	"os"

	"github.com/axelburling/dfs/cmd/master"
	"github.com/axelburling/dfs/cmd/node"
	"github.com/axelburling/dfs/cmd/pubsub"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "dfs",
	Short: "A distributed file system",
	Long: `DFS (Distributed File System) is a scalable and fault-tolerant
storage system designed to handle large-scale file storage across multiple nodes.

Features:
- Supports large file uploads with automatic chunking and replication.
- Ensures data availability with a distributed and redundant architecture.
- Provides resumable uploads for reliability in case of network failures.
- Offers efficient load balancing and dynamic node management.
- Implements a high-performance key-value store for metadata.

This CLI tool allows you to manage nodes, upload/download files, and monitor
the health of the distributed system.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	RootCmd.AddCommand(node.NodeCmd, master.MasterCmd, pubsub.PubsubCmd)

	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
}
