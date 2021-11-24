package aws

import (
	"encoding/base64"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func CreateAwsSecretSession(region string) (*secretsmanager.SecretsManager, error) {
	sess, err := CreateBaseSession(region)
	if err != nil {
		return nil, err
	}

	awsSecSvc := secretsmanager.New(sess)

	return awsSecSvc, nil
}

func CreateAwsSecret(clusterName string, clusterID string, token string, region string, logger *zap.SugaredLogger) (string, error) {
	awsSecSvc, err := CreateAwsSecretSession(region)
	if err != nil {
		return "", errors.Wrap(err, "error creating aws secret session")
	}

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

func GetAwsSecret(clusterName string, region string, logger *zap.SugaredLogger) (string, error) {
	awsSecSvc, err := CreateAwsSecretSession(region)
	if err != nil {
		return "", err
	}

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
