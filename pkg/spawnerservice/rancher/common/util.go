package common

import (
	"net/http"

	rnchrClientBase "github.com/rancher/norman/clientbase"
	rnchrClient "github.com/rancher/rancher/pkg/client/generated/management/v3"
)

const (
	username   = "token-wdjgf"
	password   = "rsh4f2b2c78m6lb5dv9kqx4p47xx5bpl4drtfnkcfczbjnb9npn282"
	rancherUrl = "https://dev-rancher-02.anatinuslabs.com/v3"
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

func CreateRancherClient() (*rnchrClient.Client, error) {
	rancherHttpClient := &http.Client{}

	// TODO: Sid add timeout
	rancherClientOpts := rnchrClientBase.ClientOpts{
		URL:        rancherUrl,
		AccessKey:  username,
		SecretKey:  password,
		HTTPClient: rancherHttpClient,
	}

	rancherClient, _ := rnchrClient.NewClient(&rancherClientOpts)

	return rancherClient, nil
}
