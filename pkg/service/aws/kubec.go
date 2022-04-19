package aws

import (
	"context"

	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookdevs/spawnerservice"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func (ctrl AWSController) GetKubeConfig(ctx context.Context, req *proto.GetKubeConfigRequest) (*proto.GetKubeConfigResponse, error) {

	region := req.Region
	clusterName := req.ClusterName

	session, err := NewSession(ctx, region, req.AccountName)
	if err != nil {
		return nil, err
	}
	client := session.getEksClient()
	ctrl.logger.Debugw("fetching cluster status", "cluster", clusterName, "region", region)

	cluster, err := getClusterSpec(ctx, client, clusterName)
	if err != nil {
		ctrl.logger.Errorw("failed to get cluster spec", "error", err, "cluster", clusterName, "region", region)
		return nil, err
	}

	kubeConfig, err := session.getKubeConfig(cluster)

	if err != nil {
		ctrl.logger.Errorw("failed to get k8s config", "error", err, "cluster", clusterName, "region", region)
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
		Exec: &clientcmdapi.ExecConfig{
			Command: "aws",
			Args: []string{
				"--region", region,
				"eks", "get-token",
				"--cluster-name", clusterName,
			},
			APIVersion: "client.authentication.k8s.io/v1alpha1",
		},
	}

	clientConfig := clientcmdapi.Config{
		Kind:       "Config",
		APIVersion: "v1",
		Clusters:   clusters,
		Contexts:   contexts,
		AuthInfos:  authinfos,
	}

	b, err := clientcmd.Write(clientConfig)
	//	clusters := []*kubeconfig.KubectlClusterWithName{
	//		&kubeconfig.KubectlClusterWithName{
	//			Name: *cluster.Arn,
	//			Cluster: kubeconfig.KubectlCluster{
	//				Server:                   kubeConfig.Host,
	//				CertificateAuthorityData: kubeConfig.CAData,
	//			},
	//		},
	//	}
	//
	//	contexts := []*kubeconfig.KubectlContextWithName{&kubeconfig.KubectlContextWithName{
	//		Name: *cluster.Arn,
	//		Context: kubeconfig.KubectlContext{
	//			Cluster: *cluster.Arn,
	//			User:    *cluster.Arn,
	//		},
	//	}}
	//
	//	users := []*kubeconfig.KubectlUserWithName{&kubeconfig.KubectlUserWithName{
	//		Name: *cluster.Arn,
	//		User: kubeconfig.KubectlUser{},
	//	}}
	//
	//	kconf := kubeconfig.KubectlConfig{
	//		Kind:       "Config",
	//		ApiVersion: "v1",
	//		Clusters:   clusters,
	//		Contexts:   contexts,
	//		Users:      users,
	//	}
	//
	//	b, err := yaml.Marshal(&kconf)
	//
	//	if err != nil {
	//		ctrl.logger.Errorw("failed to marshal kube config ", "error", err)
	//		return nil, errors.Wrap(err, "GetKubeConfig")
	//	}
	//
	//	conf, err := clientcmd.NewClientConfigFromBytes(b)
	//
	//	if err != nil {
	//		ctrl.logger.Errorw("failed to convert Config from bytes", "error", err)
	//		return nil, errors.Wrap(err, "GetKubeConfig")
	//	}
	//
	//	c, err := conf.RawConfig()
	//	if err != nil {
	//		return nil, err
	//	}
	//	spew.Dump(c)
	//
	return &proto.GetKubeConfigResponse{
		ClusterName: *cluster.Arn,
		Config:      b,
	}, nil
}

func (ctrl AWSController) GetToken(ctx context.Context, req *proto.GetTokenRequest) (*proto.GetTokenResponse, error) {

	region := req.Region
	clusterName := req.ClusterName

	session, err := NewSession(ctx, region, req.AccountName)
	if err != nil {
		return nil, err
	}
	client := session.getEksClient()
	ctrl.logger.Debugw("fetching cluster status", "cluster", clusterName, "region", region)

	cluster, err := getClusterSpec(ctx, client, clusterName)
	if err != nil {
		ctrl.logger.Errorw("failed to get cluster spec", "error", err, "cluster", clusterName, "region", region)
		return nil, err
	}

	kubeConfig, err := session.getKubeConfig(cluster)
	if err != nil {
		ctrl.logger.Errorw("failed to get k8s config", "error", err, "cluster", clusterName, "region", region)
		return nil, err
	}
	return &proto.GetTokenResponse{
		Token:    kubeConfig.BearerToken,
		CaData:   string(kubeConfig.CAData),
		Endpoint: kubeConfig.Host,
	}, nil
}
