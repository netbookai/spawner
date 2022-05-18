package service

import (
	"context"

	"gitlab.com/netbook-devs/spawner-service/pkg/service/system"
)

func (svc *spawnerService) getCredentials(ctx context.Context, region, account, credType string) (system.Credentials, error) {

	creds, err := system.GetCredentials(ctx, region, account, credType)
	if err != nil {
		svc.logger.Errorw("failed to get the credentials", "account", account)
		return nil, err
	}
	return creds, nil
}

//writeCredentials just a wrapper over system func
func (svc *spawnerService) writeCredentials(ctx context.Context, region, account, credType string, cred system.Credentials) error {

	update, err := system.WriteOrUpdateCredential(ctx, region, account, credType, cred)
	svc.logger.Infow("Secrets written successfully", "update", update)
	return err
}
