package config

import (
	"github.com/spf13/viper"
)

var config Config

// Config stores all configuration of the application.
// The values are read by viper from a config file or environment variable.
type Config struct {
	//Env value can be 'dev', 'prod' or local
	Env       string `mapstructure:"ENV"`
	Port      int    `mapstructure:"GRPC_PORT"`
	DebugPort int    `mapstructure:"HTTP_PORT"`
	//Rancher optional, requires to register cluster with rancher

	RancherUsername string `mapstructure:"RANCHER_USERNAME"`
	RancherPassword string `mapstructure:"RANCHER_PASSWORD"`
	RancherAddr     string `mapstructure:"RANCHER_ADDRESS"`

	//route 53 hosted zone id
	AwsRoute53HostedZoneID string `mapstructure:"AWS_ROUTE53_HOSTEDZONEID"`
	//Aws creds required for local runs
	AWSAccessID  string `mapstructure:"AWS_ACCESS_ID"`
	AWSSecretKey string `mapstructure:"AWS_SECRET_KEY"`
	//AWSToken optinal token for aws sessions
	AWSToken string `mapstructure:"AWS_TOKEN"`

	//SecretHostRegion aws secret manager region, used for storing user credentials
	SecretHostRegion string `mapstructure:"SECRET_HOST_REGION"`

	//NodeDeletionTimeout during the cluster deletion with force flag enabled, all attached nodes will be deleted
	//and spawner will wait till NodeDeletionTimeout before attemption cluster deletion.
	//make sure this is set sufficiently for the nodes to be deleted, otherwise cluster deletion will fail
	NodeDeletionTimeout int32 `mapstructure:"NODE_DELETION_TIME_IN_SECONDS"`

	//Azure config

	//AzureCloudProvider could be one of the following
	// [ "AZURECHINACLOUD", "AZUREGERMANCLOUD", "AZUREPUBLICCLOUD", "AZUREUSGOVERNMENTCLOUD" ]
	AzureCloudProvider  string `mapstructure:"AZURE_CLOUD_PROVIDER"`
	AzureSubscriptionID string `mapstructure:"AZURE_SUBSCRIPTION_ID"`
	AzureTenantID       string `mapstructure:"AZURE_TENANT_ID"`
	AzureClientID       string `mapstructure:"AZURE_CLIENT_ID"`
	AzureClientSecret   string `mapstructure:"AZURE_CLIENT_SECRET"`
	AzureResourceGroup  string `mapstructure:"AZURE_RESOURCE_GROUP"`
}

// Load reads configuration from file or environment variables.
func Load(path string) error {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	err = viper.Unmarshal(&config)

	return err
}

//Get retrieve cached config
func Get() Config {
	return config
}
