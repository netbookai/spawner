package cli

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/spf13/cobra"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookdevs/spawnerservice"
)

func createCluster() *cobra.Command {
	name := ""
	provider := ""
	addr := ""
	ifile := "request.json"
	c := &cobra.Command{
		Use:     "create-cluster",
		Short:   "create-cluster clustename",
		Long:    "create a cluster in given environment",
		Example: "create-cluster mycluster",
		Args: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Version:   "0.0.1",
		ValidArgs: []string{"name"},
		Run: func(cmd *cobra.Command, args []string) {
			if name == "" && len(args) < 1 {
				log.Fatal("cluster name must be provided as first argument")
			}
			if len(args) == 1 {
				name = args[0]
			}

			req := &proto.ClusterRequest{}
			data, err := os.ReadFile(ifile)
			if err != nil {
				log.Fatal("failed to read request file: ", err.Error())
			}
			err = json.Unmarshal(data, req)
			if err != nil {
				log.Fatal("failed to unmarshal request: ", err.Error())
			}
			req.ClusterName = name
			req.Provider = provider
			conn, err := getSpawnerConn(addr)
			if err != nil {
				log.Fatal("failed to connect to spawner ", addr)
			}
			defer conn.Close()
			client := proto.NewSpawnerServiceClient(conn)

			_, err = client.CreateCluster(context.Background(), req)

			//TODO: add new node as per cluster node spec
			if err != nil {
				log.Fatal("create cluster failed: ", err.Error())
			}
		},
	}
	c.Flags().StringVarP(&name, "name", "n", "", "cluster name")
	c.Flags().StringVarP(&addr, "addr", "a", "localhost:8083", "spanwner service hoost address 'ip:port'")
	c.Flags().StringVarP(&provider, "provider", "p", "aws", "cloud provider, one of ['aws', 'azure']")
	c.Flags().StringVarP(&ifile, "request", "r", "request.json", "file containing cluster spec")
	return c
}

func clusteStatus() *cobra.Command {
	name := ""
	provider := ""
	region := ""
	addr := ""

	c := &cobra.Command{
		Use:     "cluster-status",
		Short:   "cluster-status clustename",
		Long:    "Get the status of the cluster",
		Example: "cluster-status mycluster",
		Args: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Version:   "0.0.1",
		ValidArgs: []string{"name"},
		Run: func(cmd *cobra.Command, args []string) {
			if name == "" && len(args) < 1 {
				log.Fatal("cluster name must be provided as first argument or passed in as flags")
			}
			if len(args) == 1 {
				name = args[0]
			}

			req := &proto.ClusterStatusRequest{}
			req.ClusterName = name
			req.Provider = provider
			req.Region = region

			conn, err := getSpawnerConn(addr)
			if err != nil {
				log.Fatal("failed to connect to spawner ", addr)
			}
			defer conn.Close()
			client := proto.NewSpawnerServiceClient(conn)
			resp, err := client.ClusterStatus(context.Background(), req)
			if err != nil {
				log.Fatal("failed to get status: ", err.Error())
			}

			log.Println("Cluster status: ", resp.Status)
		},
	}

	c.Flags().StringVarP(&name, "name", "n", "", "cluster name")
	c.Flags().StringVarP(&provider, "provider", "p", "", "cloud provider, one of ['aws', 'azure']")
	c.Flags().StringVarP(&region, "region", "r", "", "cluster hosted region")
	c.Flags().StringVarP(&addr, "addr", "a", "localhost:8083", "spanwner service hoost address 'ip:port'")

	c.MarkFlagRequired("region")
	c.MarkFlagRequired("provider")
	return c
}

func deleteCluster() *cobra.Command {
	name := ""
	provider := ""
	region := ""
	addr := ""
	force := false

	c := &cobra.Command{
		Use:     "delete-cluster",
		Short:   "delete-cluster clustername",
		Long:    "delete existing cluster along with its node",
		Example: "delete-cluster mycluster",
		Args: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Version:   "0.0.1",
		ValidArgs: []string{"name"},
		Run: func(cmd *cobra.Command, args []string) {
			if name == "" && len(args) < 1 {
				log.Fatal("cluster name must be provided as first argument or passed in as flags")
			}
			if len(args) == 1 {
				name = args[0]
			}

			req := &proto.ClusterDeleteRequest{}
			req.ClusterName = name
			req.Provider = provider
			req.Region = region
			req.ForceDelete = force

			conn, err := getSpawnerConn(addr)
			if err != nil {
				log.Fatal("failed to connect to spawner ", addr)
			}
			defer conn.Close()
			client := proto.NewSpawnerServiceClient(conn)
			_, err = client.DeleteCluster(context.Background(), req)
			if err != nil {
				log.Fatal("failed to get status: ", err.Error())
			}

			log.Println("Cluster deleted")
		},
	}

	c.Flags().StringVarP(&name, "name", "n", "", "cluster name")
	c.Flags().StringVarP(&provider, "provider", "p", "", "cloud provider, one of ['aws', 'azure']")
	c.Flags().StringVarP(&region, "region", "r", "", "cluster hosted region")
	c.Flags().StringVarP(&addr, "addr", "a", "localhost:8083", "spanwner service hoost address 'ip:port'")
	c.Flags().BoolVarP(&force, "force", "f", false, "force delete all nodes in the cluster")

	c.MarkFlagRequired("region")
	c.MarkFlagRequired("provider")
	return c
}
