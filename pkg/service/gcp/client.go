package gcp

import (
	"context"

	container "cloud.google.com/go/container/apiv1"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/system"
	"google.golang.org/api/option"
)

func getClusterClient(ctx context.Context, cred *system.GCPCredential) (*container.ClusterManagerClient, error) {

	sa_cred := []byte(cred.Certificate)
	c, err := container.NewClusterManagerClient(ctx, option.WithCredentialsJSON(sa_cred))
	return c, err
}
