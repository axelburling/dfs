/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package pubsub

import (
	logger "github.com/axelburling/dfs/internal/log"
	"github.com/axelburling/dfs/pkg/pubsub"
	"github.com/spf13/cobra"
)

// pubsubCmd represents the pubsub command
var PubsubCmd = &cobra.Command{
	Use:   "pubsub",
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

		ps := pubsub.New(log)

		if err := ps.Start(); err != nil {
			panic(err)
		}
	},
}

func init() {
	PubsubCmd.AddCommand(subCmd)

	PubsubCmd.Flags().BoolP("dev", "d", false, "toggle node to development mode")
}
