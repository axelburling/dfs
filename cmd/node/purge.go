/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package node

import (
	logger "github.com/axelburling/dfs/internal/log"
	"github.com/axelburling/dfs/pkg/node/storage"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// purgeCmd represents the purge command
var purgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "purge data from node instance",
	Long: `The purge command removes stored file chunks from a node to free up space
or clean up orphaned data.

Options:
- Remove specific chunks or all chunks from the node.
- Purge unreferenced or corrupted data.
- Ensure consistency between the node and master metadata.

Use this command to manually clear data from a node when necessary.`,
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
			log.Fatal("failed to parse storage directory", zap.Bool("mode", dev), zap.Error(err))
		}

		store, err := storage.NewFileStorage(dir, log, nil)
		if err != nil {
			log.Fatal("failed to init storage", zap.String("path", dir), zap.Bool("mode", dev), zap.Error(err))
		}

		err = store.Purge()

		if err != nil {
			log.Fatal("failed to purge storage", zap.String("path", dir), zap.Bool("mode", dev), zap.Error(err))
		}
		log.Info("successfully purged storage", zap.String("path", dir), zap.Bool("mode", dev))
		return
	},
}

func init() {
	NodeCmd.AddCommand(purgeCmd)

	// Here you will define your flags and configuration settings.
	purgeCmd.Flags().String("dir", "data", "where to store the chunks")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	purgeCmd.Flags().BoolP("dev", "d", false, "toggle node to development mode")
}
