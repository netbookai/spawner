package cli

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookdevs/spawnerservice"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

//unmarshalFile read the given file content and umarshal it to given interface
func unmarshalFile(file string, v interface{}) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return errors.Wrap(err, "failed to read request file")
	}
	err = json.Unmarshal(data, v)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal request")
	}
	return nil
}

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

			err := unmarshalFile(ifile, req)
			if err != nil {
				log.Fatal(err.Error())
			}

			req.ClusterName = name
			if provider != "" {
				req.Provider = provider
			}
			conn, err := getSpawnerConn(addr)
			if err != nil {
				log.Fatal("failed to connect to spawner ", addr)
			}
			defer conn.Close()
			client := proto.NewSpawnerServiceClient(conn)

			log.Printf("creating cluster '%s'\n", name)

			_, err = client.CreateCluster(cmd.Context(), req)

			if err != nil {
				log.Fatal("create cluster failed: ", err.Error())
			}

			if req.Provider == "aws" {
				stat := waitUntilClusterReady(cmd.Context(), client, req)
				if stat != "ACTIVE" {
					log.Fatal("failed to wait on cluster activation, please check provider portal")
				}

				//add default node
				nsr := &proto.NodeSpawnRequest{}
				nsr.Provider = req.Provider
				nsr.Region = req.Region
				nsr.NodeSpec = req.Node
				nsr.ClusterName = req.ClusterName
				log.Printf("cluster '%s' is active, adding node '%s'\n", name, req.Node.Name)
				_, err := client.AddNode(cmd.Context(), nsr)
				if err != nil {
					log.Fatalf("failed to attach node to cluster '%s', can retry 'nodepool add' %s\n", name, err.Error())
					return
				}
				log.Println("nodepool attached to cluster")
			} else {
				log.Printf("cluster '%s' created\n", name)
			}
		},
	}
	c.Flags().StringVarP(&name, "name", "n", "", "cluster name")
	c.Flags().StringVarP(&addr, "addr", "a", "localhost:8083", "spanwner service hoost address 'ip:port'")
	c.Flags().StringVarP(&provider, "provider", "p", "", "cloud provider, one of ['aws', 'azure']")
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
			log.Printf("fetching cluster '%s' status\n", name)
			resp, err := client.ClusterStatus(cmd.Context(), req)
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
			log.Printf("deleting cluster '%s'\n", name)
			_, err = client.DeleteCluster(cmd.Context(), req)
			if err != nil {
				log.Fatal("failed to get status: ", err.Error())
			}

			log.Printf("cluster '%s' deleted\n", name)
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

func addNodePool() *cobra.Command {
	name := ""
	addr := ""
	ifile := ""

	c := &cobra.Command{
		Use:     "add",
		Short:   "add nodepool to cluster",
		Long:    "adds new nodepool to cluster",
		Example: "nodepool add",
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
			req := &proto.NodeSpawnRequest{}
			err := unmarshalFile(ifile, req)
			if err != nil {
				log.Fatal(err.Error())
			}

			req.ClusterName = name

			conn, err := getSpawnerConn(addr)
			if err != nil {
				log.Fatal("failed to connect to spawner ", addr)
			}
			defer conn.Close()
			client := proto.NewSpawnerServiceClient(conn)
			log.Printf("adding nodepool '%s' to cluster '%s'\n", req.NodeSpec.Name, name)
			_, err = client.AddNode(cmd.Context(), req)
			if err != nil {
				log.Fatal("failed to add new node pool: ", err.Error())
			}

			log.Printf("node '%s' added\n", req.NodeSpec.Name)
		},
	}

	c.Flags().StringVarP(&name, "name", "n", "", "cluster name")
	c.Flags().StringVarP(&ifile, "request", "r", "request.json", "file containing nodepool spec")
	c.Flags().StringVarP(&addr, "addr", "a", "localhost:8083", "spanwner service hoost address 'ip:port'")

	return c
}

func deleteNodePool() *cobra.Command {
	name := ""
	addr := ""
	provider := ""
	region := ""
	nodeName := ""

	c := &cobra.Command{
		Use:     "delete",
		Short:   "delete nodepool to cluster",
		Long:    "delete nodepool in the cluster",
		Example: "nodepool delete",
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
			req := &proto.NodeDeleteRequest{}

			req.ClusterName = name
			req.NodeGroupName = nodeName
			req.Provider = provider
			req.Region = region

			conn, err := getSpawnerConn(addr)
			if err != nil {
				log.Fatal("failed to connect to spawner ", addr)
			}
			defer conn.Close()
			client := proto.NewSpawnerServiceClient(conn)
			log.Printf("deleting nodepool '%s' in cluster '%s'\n", req.NodeGroupName, name)
			_, err = client.DeleteNode(cmd.Context(), req)
			if err != nil {
				log.Fatal("failed to delete node pool: ", err.Error())
			}

			log.Printf("nodepool '%s' deleted\n", req.NodeGroupName)
		},
	}

	c.Flags().StringVarP(&name, "name", "n", "", "cluster name")
	c.Flags().StringVarP(&addr, "addr", "a", "localhost:8083", "spanwner service hoost address 'ip:port'")

	c.Flags().StringVarP(&provider, "provider", "p", "", "cloud provider, one of ['aws', 'azure']")
	c.Flags().StringVarP(&region, "region", "r", "", "cluster hosted region")
	c.Flags().StringVar(&nodeName, "nodepool", "", "nodepool to be deleted")

	c.MarkFlagRequired("nodepool")
	c.MarkFlagRequired("region")
	c.MarkFlagRequired("provider")

	return c
}

func nodepool() *cobra.Command {

	c := &cobra.Command{
		Use:   "nodepool",
		Short: "nodepool [add|delete]",
		Long:  "add or delete nodepool from cluster",
	}
	c.AddCommand(addNodePool())
	c.AddCommand(deleteNodePool())
	return c
}

func kubeConfig() *cobra.Command {
	name := ""
	addr := ""
	provider := ""
	region := ""

	c := &cobra.Command{

		Use:     "kubeconfig",
		Short:   "get kubeconfig for the cluster",
		Long:    "get kubeconfig for the cluster",
		Example: "kubeconfig clustername",
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
			req := &proto.GetKubeConfigRequest{}

			req.ClusterName = name
			req.Provider = provider
			req.Region = region

			conn, err := getSpawnerConn(addr)
			if err != nil {
				log.Fatal("failed to connect to spawner ", addr)
			}
			defer conn.Close()
			client := proto.NewSpawnerServiceClient(conn)
			log.Printf("getting kube config for the cluster '%s'\n", name)
			res, err := client.GetKubeConfig(cmd.Context(), req)
			if err != nil {
				log.Fatalf("failed to get kube config: %s\n", err.Error())
			}

			newConfig, err := clientcmd.Load(res.GetConfig())
			if err != nil {
				log.Fatalf("failed to read kube config : %s\n", err.Error())
			}

			home, err := os.UserHomeDir()
			if err != nil {
				log.Fatalf("failed to read user home directory: %s\n", err.Error())
			}

			kubefile := fmt.Sprintf("%s/.kube/config", home)
			log.Printf("reading existing kube config : %s\n", kubefile)

			skipMerge := false
			currentKC, err := clientcmd.LoadFromFile(kubefile)

			if os.IsNotExist(err) {
				//file does not exist, create new one
				skipMerge = true

			}

			if !skipMerge && err != nil {
				log.Fatalf("failed to load the existing kube config : %s\n", err.Error())
			}

			kubeConfg := clientcmdapi.NewConfig()
			//set the current cluster context as new context
			newConfig.CurrentContext = res.ClusterName
			if !skipMerge {
				mergo.Merge(kubeConfg, currentKC, mergo.WithOverride)
			}
			mergo.Merge(kubeConfg, newConfig, mergo.WithOverride)

			err = clientcmd.WriteToFile(*kubeConfg, kubefile)
			if err != nil {
				log.Fatalf("failed to write kube config : %s\n", err.Error())
			}
			log.Printf("KubeConfig updated in '%s'\n", kubefile)
		},
	}

	c.Flags().StringVarP(&name, "name", "n", "", "cluster name")
	c.Flags().StringVarP(&addr, "addr", "a", "localhost:8083", "spanwner service hoost address 'ip:port'")

	c.Flags().StringVarP(&provider, "provider", "p", "", "cloud provider, one of ['aws', 'azure']")
	c.Flags().StringVarP(&region, "region", "r", "", "cluster hosted region")

	return c
}
