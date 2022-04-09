package system

import (
	"errors"
	"fmt"
	"strings"
)

type AzureCredential struct {
	SubscriptionID string
	TenantID       string
	ClientID       string
	ClientSecret   string
	ResourceGroup  string
	Name           string
}

type AwsCredential struct {
	Name   string
	Id     string
	Secret string
	Token  string
}

type Credentials interface {
	GetAzure() *AzureCredential
	GetAws() *AwsCredential
	AsSecretValue() string
}

var _ Credentials = (*AzureCredential)(nil)
var _ Credentials = (*AwsCredential)(nil)

//Azure credentials

func (a *AzureCredential) GetAzure() *AzureCredential {
	return a
}

func (a *AzureCredential) GetAws() *AwsCredential {
	return nil
}

func (a *AzureCredential) AsSecretValue() string {
	return fmt.Sprintf("%s,%s,%s,%s,%s", a.SubscriptionID, a.TenantID, a.ClientID, a.ClientSecret, a.ResourceGroup)
}

//Aws credential

func (a *AwsCredential) GetAzure() *AzureCredential {
	return nil
}

func (a *AwsCredential) GetAws() *AwsCredential {
	return a
}

func (a *AwsCredential) AsSecretValue() string {
	return fmt.Sprintf("%s,%s,%s", a.Id, a.Secret, a.Token)
}

func NewAwsCredential(blob string) (*AwsCredential, error) {
	//secret_id,secret,token
	splits := strings.Split(blob, ",")
	if len(splits) == 3 {
		return &AwsCredential{
			Id:     splits[0],
			Secret: splits[1],
			Token:  splits[2],
		}, nil
	}
	if len(splits) == 2 {
		//older format where we ignored token
		return &AwsCredential{
			Id:     splits[0],
			Secret: splits[1],
		}, nil
	}
	return nil, errors.New("NewAwsCredential: invalid credentials found in secrets")
}

func NewAzureCredential(blob string) (*AzureCredential, error) {
	splits := strings.Split(blob, ",")
	if len(splits) != 5 {
		return nil, errors.New("NewAzureCredential: invalid credentials found in secrets")
	}
	return &AzureCredential{
		SubscriptionID: splits[0],
		TenantID:       splits[1],
		ClientID:       splits[2],
		ClientSecret:   splits[3],
		ResourceGroup:  splits[4],
	}, nil

}
