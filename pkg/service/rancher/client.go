package rancher

import (
	"fmt"
	"net/http"

	rnchrClientBase "github.com/rancher/norman/clientbase"
	rnchrClient "github.com/rancher/rancher/pkg/client/generated/management/v3"
)

func CreateRancherClient(url string, accessKey string, secretKey string) (*rnchrClient.Client, error) {
	rancherHttpClient := &http.Client{}

	// TODO: Sid add timeout
	rancherClientOpts := rnchrClientBase.ClientOpts{
		URL:        url,
		AccessKey:  accessKey,
		SecretKey:  secretKey,
		HTTPClient: rancherHttpClient,
	}

	rancherClient, err := rnchrClient.NewClient(&rancherClientOpts)
	if err != nil {
		fmt.Println(fmt.Errorf("error creating rancher client %s", err))
		return rancherClient, err
	}
	return rancherClient, nil
}
