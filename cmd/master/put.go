/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package master

import (
	"os"

	logger "github.com/axelburling/dfs/internal/log"
	"github.com/axelburling/dfs/pkg/master/queue"
	"github.com/axelburling/dfs/pkg/node/client"
	"github.com/spf13/cobra"
)

// putCmd represents the master command
var putCmd = &cobra.Command{
	Use:   "put",
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

		gq := queue.NewGlobalQueue(log)

		node, err := client.NewNode("http://localhost:4000", "localhost:4001")

		if err != nil {
			panic(err)
		}

		gq.AddNode(node)

		f, err := os.Open("sqlc.yaml")

		if err != nil {
			panic(err)
		}

		gq.Enqueue(queue.NewJob(f, 0, 1))

		// ids := make([]string, 0)
		// nodes := make([]*client.Node, 0)

		// for range 10 {
		// 	id := uuid.NewString()
		// 	node := &client.Node{
		// 		ID: id,
		// 	}
		// 	ids = append(ids, id)
		// 	nodes = append(nodes, node)
		// 	gq.AddNode(node)
		// }

		// gq.Enqueue(queue.Job{
		// 	ChunkIndex: int64(rand.Intn(100)),
		// 	ID: uuid.NewString(),
		// 	NodeIds: []string{ids[0]},

		// })
		// time.Sleep(500*time.Millisecond)

		for {
		}
		// for {
		// 	time.Sleep(400*time.Millisecond)
		// }
	},
}

func init() {
	MasterCmd.AddCommand(putCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// masterCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// masterCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	putCmd.Flags().BoolP("dev", "d", false, "toggle node to production")
}
