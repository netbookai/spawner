package gcp

import (
	"context"

	compute "cloud.google.com/go/compute/apiv1"
	container "cloud.google.com/go/container/apiv1"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/system"
	auth "golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

func getClusterManagerClient(ctx context.Context, cred *system.GCPCredential) (*container.ClusterManagerClient, error) {

	sa_cred := []byte(cred.Certificate)
	opt := option.WithCredentialsJSON(sa_cred)
	c, err := container.NewClusterManagerClient(ctx, opt)
	return c, err
}

//getDiskClient
func getDiskClient(ctx context.Context, cred *system.GCPCredential) (*compute.DisksClient, error) {

	sa_cred := []byte(cred.Certificate)
	opt := option.WithCredentialsJSON(sa_cred)
	return compute.NewDisksRESTClient(ctx, opt)
}

//getSnapshotClient
func getSnapshotClient(ctx context.Context, cred *system.GCPCredential) (*compute.SnapshotsClient, error) {

	sa_cred := []byte(cred.Certificate)
	opt := option.WithCredentialsJSON(sa_cred)
	return compute.NewSnapshotsRESTClient(ctx, opt)
}

//getAuthClient
func getAuthClient(ctx context.Context, cred *system.GCPCredential) (*auth.Credentials, error) {

	sa_cred := []byte(cred.Certificate)
	scopes := []string{
		"https://www.googleapis.com/auth/cloud-platform",
	}
	return auth.CredentialsFromJSON(ctx, sa_cred, scopes...)
}
