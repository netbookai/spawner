package gcp

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
	"google.golang.org/genproto/googleapis/container/v1"
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

	auth, err := getAuthClient(ctx, cred)
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
	return nil, nil
}
