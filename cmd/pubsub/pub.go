/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package pubsub

import (
	"context"
	"fmt"

	"github.com/axelburling/dfs/internal/config"
	logger "github.com/axelburling/dfs/internal/log"
	"github.com/axelburling/dfs/pkg/pubsub/client/apiv1/message"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// pubCmd represents the pub command
var pubCmd = &cobra.Command{
	Use:   "publish",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
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

		topicName, err := cmd.Flags().GetString("topic")
		if err != nil {
			panic(err)
		}

		action, err := cmd.Flags().GetString("action")
		if err != nil {
			panic(err)
		}

		data, err := cmd.Flags().GetString("data")
		if err != nil {
			panic(err)
		}

		conf := config.New[config.PubSubConfig](config.PubSub, log)

		pubsub := setupPubSub(log, conf)

		topic, err := pubsub.CreateTopic(ctx, topicName)
		if err != nil {
			panic(err)
		}

		fmt.Println(topic.Name)

		msg := topic.NewMessage(action, []byte(data))

		ids, err := topic.Publish(ctx, []*message.Message{msg}, nil)
		if err != nil {
			panic(err)
		}

		log.Info("published message", zap.Strings("ids", ids))
	},
}

func init() {
	PubsubCmd.AddCommand(pubCmd)

	pubCmd.Flags().BoolP("dev", "d", false, "toggle node to development mode")
	pubCmd.Flags().StringP("topic", "t", "test", "toggle node to development mode")
	pubCmd.Flags().StringP("action", "a", "node:add", "toggle node to development mode")
	pubCmd.Flags().String("data", "id:ndbknbfdbbdf", "toggle node to development mode")
}
