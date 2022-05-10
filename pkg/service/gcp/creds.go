package gcp

import (
	"context"

	"github.com/pkg/errors"
	"gitlab.com/netbook-devs/spawner-service/pkg/config"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/constants"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/system"
)

func getCredentials(ctx context.Context, account string) (*system.GCPCredential, error) {
	env := config.Get().Env

	if env == "local" {
		conf := config.Get()
		return &system.GCPCredential{
			ProjectId:   conf.GcpProject,
			Certificate: conf.GcpCertificate,
		}, nil
	} else {
		c, err := system.GetCredentials(ctx, config.Get().SecretHostRegion, account, constants.GcpLabel)
		if err != nil {
			return nil, errors.Wrap(err, "getCredentials")
		}
		return c.GetGcp(), nil
	}
}
