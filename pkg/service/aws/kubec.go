package aws

import (
	"context"

	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

//GetKubeConfig generates the kubeconfig from the session
func (ctrl AWSController) GetKubeConfig(ctx context.Context, req *proto.GetKubeConfigRequest) (*proto.GetKubeConfigResponse, error) {

	region := req.Region
	clusterName := req.ClusterName

	session, err := NewSession(ctx, region, req.AccountName)
	if err != nil {
		return nil, err
	}
	client := session.getEksClient()
	ctrl.logger.Debug(ctx, "fetching cluster status", "cluster", clusterName, "region", region)

	cluster, err := getClusterSpec(ctx, client, clusterName)
	if err != nil {
		ctrl.logger.Error(ctx, "failed to get cluster spec", "error", err, "cluster", clusterName, "region", region)
		return nil, err
	}

	kubeConfig, err := session.getKubeConfig(cluster)

	if err != nil {
		ctrl.logger.Error(ctx, "failed to get k8s config", "error", err, "cluster", clusterName, "region", region)
		return nil, err
	}

	defaultCluster := *cluster.Arn

	clusters := make(map[string]*clientcmdapi.Cluster)
	clusters[defaultCluster] = &clientcmdapi.Cluster{
		Server:                   kubeConfig.Host,
		CertificateAuthorityData: kubeConfig.CAData,
	}

	contexts := make(map[string]*clientcmdapi.Context)
	contexts[defaultCluster] = &clientcmdapi.Context{
		Cluster:  defaultCluster,
		AuthInfo: defaultCluster,
	}

	authinfos := make(map[string]*clientcmdapi.AuthInfo)
	authinfos[defaultCluster] = &clientcmdapi.AuthInfo{
		Token: kubeConfig.BearerToken,
	}

	clientConfig := clientcmdapi.Config{
		Kind:       "Config",
		APIVersion: "v1",
		Clusters:   clusters,
		Contexts:   contexts,
		AuthInfos:  authinfos,
	}

	b, err := clientcmd.Write(clientConfig)
	return &proto.GetKubeConfigResponse{
		ClusterName: *cluster.Arn,
		Config:      b,
	}, nil
}

//GetToken get aws tokens and ca data for kube
func (ctrl AWSController) GetToken(ctx context.Context, req *proto.GetTokenRequest) (*proto.GetTokenResponse, error) {

	region := req.Region
	clusterName := req.ClusterName

	session, err := NewSession(ctx, region, req.AccountName)
	if err != nil {
		return nil, err
	}
	client := session.getEksClient()
	ctrl.logger.Debug(ctx, "fetching cluster status", "cluster", clusterName, "region", region)

	cluster, err := getClusterSpec(ctx, client, clusterName)
	if err != nil {
		ctrl.logger.Error(ctx, "failed to get cluster spec", "error", err, "cluster", clusterName, "region", region)
		return nil, err
	}

	kubeConfig, err := session.getKubeConfig(cluster)
	if err != nil {
		ctrl.logger.Error(ctx, "failed to get k8s config", "error", err, "cluster", clusterName, "region", region)
		return nil, err
	}
	return &proto.GetTokenResponse{
		Token:    kubeConfig.BearerToken,
		CaData:   string(kubeConfig.CAData),
		Endpoint: kubeConfig.Host,
	}, nil
}
