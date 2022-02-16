package config

import (
	"github.com/spf13/viper"
)

// Config stores all configuration of the application.
// The values are read by viper from a config file or environment variable.
type Config struct {
	//enviroment is either development or production
	Env                    string `mapstructure:"ENV"`
	Port                   int    `mapstructure:"PORT"`
	DebugPort              int    `mapstructure:"DEBUG_PORT"`
	RancherUsername        string `mapstructure:"RANCHER_USERNAME"`
	RancherPassword        string `mapstructure:"RANCHER_PASSWORD"`
	RancherAddr            string `mapstructure:"RANCHER_ADDRESS"`
	AwsCredName            string `mapstructure:"RANCHER_AWS_CRED_NAME"`
	AwsRoute53HostedZoneID string `mapstructure:"AWS_ROUTE53_HOSTEDZONEID"`
	AWSAccessID            string `mapstructure:"AWS_ACCESS_ID"`
	AWSSecretKey           string `mapstructure:"AWS_SECRET_KEY"`
	AWSToken               string `mapstructure:"AWS_TOKEN"`
}

// Load reads configuration from file or environment variables.
func Load(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
