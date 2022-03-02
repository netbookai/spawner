package spawnerservice

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice/system"
)

func (svc *SpawnerService) getCredentials(ctx context.Context, region, account string) (credentials.Value, error) {

	creds, err := system.GetAwsCredentials(ctx, region, account)
	if err != nil {
		svc.logger.Errorw("failed to get the credentials", "account", account)
		return credentials.Value{}, err
	}
	return creds.Get()
}

//writeCredentials just a wrapper over system func
func (svc *SpawnerService) writeCredentials(ctx context.Context, region, account, id, key string) error {

	return system.WriteCredential(ctx, region, account, id, key)
}
