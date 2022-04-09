package azure

import (
	"bytes"
	"context"

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2022-01-01/containerservice"
	"github.com/pkg/errors"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookdevs/spawnerservice"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/kops/pkg/kubeconfig"
)

func (a *AzureController) getToken(ctx context.Context, req *proto.GetTokenRequest) (*proto.GetTokenResponse, error) {

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
		a.logger.Errorw("failed to get kube config", "error", err)
		return nil, err
	}

	if len(*res.Kubeconfigs) > 1 {
		a.logger.Warnw("got kube config", len(*res.Kubeconfigs))
	}
	kc := (*res.Kubeconfigs)[0]
	var kconf kubeconfig.KubectlConfig

	err = yaml.NewYAMLOrJSONDecoder(bytes.NewReader(*kc.Value), 2048).Decode(&kconf)
	if err != nil {
		a.logger.Errorw("failed to unmarshall kube file", "error", err)
	}

	//TODO: verify that multiple users wont be present when we make kube config query

	return &proto.GetTokenResponse{
		Token:    kconf.Users[0].User.Token,
		Endpoint: kconf.Clusters[0].Cluster.Server,
		Status:   "",
		CaData:   string(kconf.Clusters[0].Cluster.CertificateAuthorityData),
	}, nil
}
