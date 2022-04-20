package cli

import (
	"context"
	"log"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var rootCommand = &cobra.Command{
	Use:   "spawner",
	Short: "spawner",
	Long:  "cli to interact with slef hosted spawner service",
}

func getSpawnerConn(addr string) (*grpc.ClientConn, error) {
	log.Println("connecting to ", addr, "...")
	return grpc.Dial(addr, grpc.WithInsecure(), grpc.WithTimeout(time.Second))
}

func setupCommands() {
	rootCommand.AddCommand(createCluster())
	rootCommand.AddCommand(clusteStatus())
	rootCommand.AddCommand(deleteCluster())
	rootCommand.AddCommand(nodepool())
	rootCommand.AddCommand(kubeConfig())
}

//Execute sets up a command execute command handlers
func Execute() error {
	setupCommands()
	return rootCommand.ExecuteContext(context.Background())
}
