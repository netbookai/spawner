package cli

import (
	"log"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var rootCommand = &cobra.Command{
	Use:   "spawner-cli",
	Short: "spawner-cli",
	Long:  "cli to interact with slef hosted spawner service",
}

func getSpawnerConn(addr string) (*grpc.ClientConn, error) {

	log.Println("connecting to ", addr)
	return grpc.Dial(addr, grpc.WithInsecure(), grpc.WithTimeout(time.Second))
}

func setupCommands(l *zap.SugaredLogger) {
	rootCommand.AddCommand(createCluster())
	rootCommand.AddCommand(clusteStatus())
	rootCommand.AddCommand(deleteCluster())
}

func Execute() error {
	z, err := zap.NewDevelopment()
	if err != nil {
		return err
	}
	logger := z.Sugar()
	setupCommands(logger)
	return rootCommand.Execute()
}
