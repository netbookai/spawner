package gcp

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
	"google.golang.org/genproto/googleapis/container/v1"
	"k8s.io/client-go/tools/clientcmd"

	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func (g *GCPController) getToken(ctx context.Context, req *proto.GetTokenRequest) (*proto.GetTokenResponse, error) {

	cred, err := getCredentials(ctx, req.AccountName)
	if err != nil {
		return nil, errors.Wrap(err, "getCluster:")
	}
	cluster, err := g.getClusterInternal(ctx, cred, req.Region, req.ClusterName)

	if err != nil {
		return nil, err
	}

	if cluster.Status != container.Cluster_RUNNING {
		return nil, fmt.Errorf("cluster is not in running state yet")
	}
	server := fmt.Sprintf("https://%s", cluster.Endpoint)

	g.logger.Info(ctx, "cluster config", "server", server)

	scopes := []string{}
	nodePools := cluster.NodePools

	if len(nodePools) > 0 {
		scopes = append(scopes, nodePools[0].Config.OauthScopes...)
	}

	auth, err := getAuthClient(ctx, cred, scopes)
	if err != nil {
		g.logger.Error(ctx, "failed to get auth2 client", "error", err)
		return nil, err
	}

	t, err := auth.TokenSource.Token()
	if err != nil {
		return nil, err
	}
	token := t.AccessToken
	return &proto.GetTokenResponse{
		Token:    token,
		Endpoint: server,
		CaData:   cluster.MasterAuth.GetClusterCaCertificate(),
	}, nil
}

func (g *GCPController) getKubeConfig(ctx context.Context, req *proto.GetKubeConfigRequest) (*proto.GetKubeConfigResponse, error) {

	cred, err := getCredentials(ctx, req.AccountName)
	if err != nil {
		return nil, errors.Wrap(err, "getCluster:")
	}
	cluster, err := g.getClusterInternal(ctx, cred, req.Region, req.ClusterName)

	if err != nil {
		return nil, err
	}

	if cluster.Status != container.Cluster_RUNNING {
		return nil, fmt.Errorf("cluster is not in running state yet")
	}
	server := fmt.Sprintf("https://%s", cluster.Endpoint)

	g.logger.Info(ctx, "cluster config", "server", server)

	scopes := []string{}
	nodePools := cluster.NodePools

	if len(nodePools) > 0 {
		scopes = append(scopes, nodePools[0].Config.OauthScopes...)
	}

	auth, err := getAuthClient(ctx, cred, scopes)
	if err != nil {
		g.logger.Error(ctx, "failed to get auth2 client", "error", err)
		return nil, err
	}

	t, err := auth.TokenSource.Token()
	if err != nil {
		return nil, err
	}
	token := t.AccessToken
	name := cluster.Name

	defaultCluster := name

	clusters := make(map[string]*clientcmdapi.Cluster)
	clusters[defaultCluster] = &clientcmdapi.Cluster{
		Server:                   server,
		CertificateAuthorityData: []byte(cluster.MasterAuth.GetClusterCaCertificate()),
	}

	contexts := make(map[string]*clientcmdapi.Context)
	contexts[defaultCluster] = &clientcmdapi.Context{
		Cluster:  defaultCluster,
		AuthInfo: defaultCluster,
	}

	authinfos := make(map[string]*clientcmdapi.AuthInfo)
	if req.RawToken {
		authinfos[defaultCluster] = &clientcmdapi.AuthInfo{
			Token: token,
		}
	} else {
		authinfos[defaultCluster] = &clientcmdapi.AuthInfo{
			AuthProvider: &clientcmdapi.AuthProviderConfig{
				Name: "gcp",
				Config: map[string]string{
					"cmd-args":   "config config-helper --format=json",
					"cmd-path":   "gcloud",
					"expiry-key": "{.credential.token_expiry}",
					"token-key":  "{.credential.access_token}",
				},
			},
		}
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
		ClusterName: name,
		Config:      b,
	}, nil
}
