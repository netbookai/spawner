package common

import (
	"fmt"
	"net/http"

	rnchrClientBase "github.com/rancher/norman/clientbase"
	rnchrClient "github.com/rancher/rancher/pkg/client/generated/management/v3"
)

func Int64Ptr(i int64) *int64 {
	return &i
}

func StrPtr(s string) *string {
	return &s
}

func BoolPtr(b bool) *bool {
	return &b
}

func MapPtr(b map[string]string) *map[string]string {
	return &b
}

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
