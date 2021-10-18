package aws

import (
	"encoding/base64"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

func CreateAwsSecretSession(provider string, region string) (awsSecSvc *secretsmanager.SecretsManager) {
	//starts an AWS session

	sess, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	awsSecSvc = secretsmanager.New(sess)
	if err != nil {
		log.Fatalf("error starting aws session")
	}
	return awsSecSvc
}

func CreateAwsSecret(clusterName string, clusterID string, token string, region string) (string, error) {

	awsSecSvc := CreateAwsSecretSession("aws", region)

	encodedToken := base64.StdEncoding.EncodeToString([]byte(token))
	encodedClusterID := base64.StdEncoding.EncodeToString([]byte(clusterID))
	descriptionString := "Access token for cluster " + clusterName
	secretStringValue := "{\"ClusterID\":\"" + encodedClusterID + "\",\"BearerToken\":\"" + encodedToken + "\"}"

	input := &secretsmanager.CreateSecretInput{
		Description:  aws.String(descriptionString),
		Name:         aws.String(clusterName),
		SecretString: aws.String(secretStringValue),
	}

	result, err := awsSecSvc.CreateSecret(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			return result.String(), aerr
		}
	}
	return result.String(), err
}

func GetAwsSecret(clusterName string, region string) (string, error) {

	awsSecSvc := CreateAwsSecretSession("aws", region)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(clusterName),
		VersionStage: aws.String("AWSCURRENT"),
	}

	result, err := awsSecSvc.GetSecretValue(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {

			return result.String(), aerr
		}

	}
	return *result.SecretString, err
}
