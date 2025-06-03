/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package pubsub

import (
	"context"

	"github.com/axelburling/dfs/internal/config"
	logger "github.com/axelburling/dfs/internal/log"
	"github.com/axelburling/dfs/pkg/pubsub/client/apiv1/message"
	ms "github.com/axelburling/dfs/pkg/pubsub/client/msg"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// subCmd represents the sub command
var subCmd = &cobra.Command{
	Use:   "subscribe",
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

		subName, err := cmd.Flags().GetString("subscription")
		if err != nil {
			panic(err)
		}

		filters, err := cmd.Flags().GetStringArray("filters")

		if err != nil {
			panic(err)
		}

		conf := config.New[config.PubSubConfig](config.PubSub, log)

		pubsub := setupPubSub(log, conf)

		topic, err := pubsub.CreateTopic(ctx, topicName)
		if err != nil {
			panic(err)
		}

		sub, err := topic.CreateSubscription(ctx, subName)
		if err != nil {
			panic(err)
		}

		messageChan := make(chan *message.Message)

		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case msg := <-messageChan:

					if ms.Action(msg.Action) == ms.NodeAdd {
						data, err := ms.Decode[ms.NodeAddReq](msg)

						if err != nil {
							log.Warn("could not decode message", zap.Error(err))
						}

						log.Info("received message", zap.String("id", msg.ID), zap.String("action", msg.Action), zap.Any("data", data))

					}

					log.Info("received message", zap.String("id", msg.ID), zap.String("action", msg.Action), zap.String("data", string(msg.Data)))

					msg.Ack()
				}
			}
		}()

		log.Info("listining for new messages", zap.String("topic", topic.Name), zap.String("subscription", sub.Name), zap.Strings("filters", filters))

		err = sub.Receive(ctx, messageChan, filters)
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	subCmd.Flags().BoolP("dev", "d", false, "toggle node to development mode")
	subCmd.Flags().StringP("topic", "t", "test", "toggle node to development mode")
	subCmd.Flags().StringP("subscription", "s", "sub", "toggle node to development mode")
	subCmd.Flags().StringArrayP("filters", "f", []string{}, "")
}
