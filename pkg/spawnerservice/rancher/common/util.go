package common

import (
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

func CreateRancherClient(url string, accessKey string, secretKey string) (*rnchrClient.Client, error) {
	rancherHttpClient := &http.Client{}

	// TODO: Sid add timeout
	rancherClientOpts := rnchrClientBase.ClientOpts{
		URL:        url,
		AccessKey:  accessKey,
		SecretKey:  secretKey,
		HTTPClient: rancherHttpClient,
	}

	rancherClient, _ := rnchrClient.NewClient(&rancherClientOpts)

	return rancherClient, nil
}
