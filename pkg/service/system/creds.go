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

<<<<<<< HEAD
type GithubPersonalAccessToken struct {
	Name  string
	Token string
}

||||||| parent of c438c6f (implement get clusters, cluster status, delete cluster, credentials for gcp)
=======
type GCPCredential struct {
	Name        string
	ProjectId   string
	Certificate string
}

>>>>>>> c438c6f (implement get clusters, cluster status, delete cluster, credentials for gcp)
type Credentials interface {
	GetAzure() *AzureCredential
	GetAws() *AwsCredential
<<<<<<< HEAD
	GetGitPAT() *GithubPersonalAccessToken
||||||| parent of c438c6f (implement get clusters, cluster status, delete cluster, credentials for gcp)
=======
	GetGcp() *GCPCredential
>>>>>>> c438c6f (implement get clusters, cluster status, delete cluster, credentials for gcp)
	AsSecretValue() string
}

var _ Credentials = (*AzureCredential)(nil)
var _ Credentials = (*AwsCredential)(nil)
<<<<<<< HEAD
var _ Credentials = (*GithubPersonalAccessToken)(nil)
||||||| parent of c438c6f (implement get clusters, cluster status, delete cluster, credentials for gcp)
=======
var _ Credentials = (*GCPCredential)(nil)
>>>>>>> c438c6f (implement get clusters, cluster status, delete cluster, credentials for gcp)

//Azure credentials

func (a *AzureCredential) GetAzure() *AzureCredential {
	return a
}

func (a *AzureCredential) GetAws() *AwsCredential {
	return nil
}

func (a *AzureCredential) GetGcp() *GCPCredential {
	return nil
}

func (a *AzureCredential) AsSecretValue() string {
	return fmt.Sprintf("%s,%s,%s,%s,%s", a.SubscriptionID, a.TenantID, a.ClientID, a.ClientSecret, a.ResourceGroup)
}

func (a *AzureCredential) GetGitPAT() *GithubPersonalAccessToken {
	return nil
}

//Aws credential

func (a *AwsCredential) GetAzure() *AzureCredential {
	return nil
}

func (a *AwsCredential) GetAws() *AwsCredential {
	return a
}

func (a *AwsCredential) GetGcp() *GCPCredential {
	return nil
}

func (a *AwsCredential) AsSecretValue() string {
	return fmt.Sprintf("%s,%s,%s", a.Id, a.Secret, a.Token)
}

<<<<<<< HEAD
func (a *AwsCredential) GetGitPAT() *GithubPersonalAccessToken {
	return nil
}

//GithubPersonalAccessToken credential

func (g *GithubPersonalAccessToken) GetGitPAT() *GithubPersonalAccessToken {
	return g
}

func (g *GithubPersonalAccessToken) GetAzure() *AzureCredential {
	return nil
}

func (g *GithubPersonalAccessToken) GetAws() *AwsCredential {
	return nil
}

func (g *GithubPersonalAccessToken) AsSecretValue() string {
	return fmt.Sprintf("%s", g.Token)
}

||||||| parent of c438c6f (implement get clusters, cluster status, delete cluster, credentials for gcp)
=======
//GCP credentials, of service account

func (g *GCPCredential) GetAzure() *AzureCredential {
	return nil
}

func (g *GCPCredential) GetAws() *AwsCredential {
	return nil
}

func (g *GCPCredential) GetGcp() *GCPCredential {
	return g
}

func (g *GCPCredential) AsSecretValue() string {
	return fmt.Sprintf("%s,%s", g.ProjectId, g.Certificate)
}

>>>>>>> c438c6f (implement get clusters, cluster status, delete cluster, credentials for gcp)
//NewAwsCredential recieves comma separated list of credential parts and creates a AwsCredential
//there can be 2 or 3 parts, when the token is present we will use the latest version of Credentials
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

//NewAzureCredential recieves comma separated list of credential parts and creates a AzureCredential
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
<<<<<<< HEAD

//NewGitPAT return new GithubPersonalAccessToken
func NewGitPAT(blob string) (*GithubPersonalAccessToken, error) {
	return &GithubPersonalAccessToken{Token: blob}, nil
}
||||||| parent of c438c6f (implement get clusters, cluster status, delete cluster, credentials for gcp)
=======

func NewGcpCredential(blob string) (*GCPCredential, error) {
	splits := strings.Split(blob, ",")
	if len(splits) != 2 {
		return nil, errors.New("NewAzureCredential: invalid credentials found in secrets")
	}
	return &GCPCredential{
		ProjectId:   splits[0],
		Certificate: splits[1],
	}, nil
}
>>>>>>> c438c6f (implement get clusters, cluster status, delete cluster, credentials for gcp)
