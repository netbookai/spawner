package azure

import (
	"bytes"
	"context"

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2022-01-01/containerservice"
	"github.com/pkg/errors"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/kops/pkg/kubeconfig"
)

func (a *AzureController) kubeConfig(ctx context.Context, req *proto.GetKubeConfigRequest) ([]byte, error) {

	account := req.AccountName

	cred, err := getCredentials(ctx, account)
	if err != nil {
		return nil, err
	}
	aksClient, err := getAKSClient(cred)
	if err != nil {
		return nil, errors.Wrap(err, "creaetAKSCluster: cannot to get AKS client")
	}

	clusterName := req.ClusterName
	groupName := cred.ResourceGroup

	fqdn := ""
	res, err :=
		aksClient.ListClusterUserCredentials(ctx, groupName, clusterName, fqdn, containerservice.FormatAzure)

	if err != nil {
		a.logger.Error(ctx, "failed to get kube config", "error", err)
		return nil, err
	}

	if len(*res.Kubeconfigs) > 1 {
		a.logger.Warn(ctx, "got kube config", len(*res.Kubeconfigs))
	}
	return *(*res.Kubeconfigs)[0].Value, nil
}

func (a *AzureController) getToken(ctx context.Context, req *proto.GetTokenRequest) (*proto.GetTokenResponse, error) {

	kc, err := a.kubeConfig(ctx, &proto.GetKubeConfigRequest{
		Provider:    req.Provider,
		Region:      req.Region,
		AccountName: req.AccountName,
		ClusterName: req.ClusterName,
	})

	if err != nil {
		a.logger.Error(ctx, "failed to get kube config", "error", err)
		return nil, errors.Wrap(err, "getToken: failed to get kube config")
	}

	//parse into kubeconfig struct
	var kconf kubeconfig.KubectlConfig
	err = yaml.NewYAMLToJSONDecoder(bytes.NewReader(kc)).Decode(&kconf)
	if err != nil {
		a.logger.Error(ctx, "failed to unmarshall kube file", "error", err)
		return nil, errors.Wrap(err, "getToken: ")
	}

	//TODO: verify that multiple users wont be present when we make kube config query

	return &proto.GetTokenResponse{
		Token:    kconf.Users[0].User.Token,
		Endpoint: kconf.Clusters[0].Cluster.Server,
		Status:   "",
		CaData:   string(kconf.Clusters[0].Cluster.CertificateAuthorityData),
	}, nil
}

func (a *AzureController) getKubeConfig(ctx context.Context, req *proto.GetKubeConfigRequest) (*proto.GetKubeConfigResponse, error) {
	kc, err := a.kubeConfig(ctx, req)
	if err != nil {
		a.logger.Error(ctx, "failed to get kube config", "error", err)
		return nil, errors.Wrap(err, "getKubeConfig: failed to get kube config")
	}

	return &proto.GetKubeConfigResponse{
		ClusterName: req.ClusterName,
		Config:      kc,
	}, nil
}
