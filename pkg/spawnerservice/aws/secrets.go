package aws

import (
	"encoding/base64"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/sts"
	"go.uber.org/zap"
)

func GetCredsFromSTS(logger *zap.SugaredLogger) (string, string, string, error) {
	svc := sts.New(session.New())
	web_identity_token, err := os.ReadFile("/var/run/secrets/eks.amazonaws.com/serviceaccount/token")
	if err != nil {
		logger.Errorw("error reading token", "error", err)
	}
	input := &sts.AssumeRoleWithWebIdentityInput{
		DurationSeconds:  aws.Int64(900),
		RoleArn:          aws.String(os.Getenv("AWS_ROLE_ARN")),
		RoleSessionName:  aws.String("SecretsConnection"),
		WebIdentityToken: aws.String(string(web_identity_token)),
	}

	result, err := svc.AssumeRoleWithWebIdentity(input)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case sts.ErrCodeExpiredTokenException:
				logger.Errorw("token expired", "error", aerr.Error())
			case sts.ErrCodeRegionDisabledException:
				logger.Errorw("error creating token: region disabled", "error", aerr.Error())
			default:
				logger.Errorw("error accessing aws", "error", aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			logger.Errorw("error while getting credentials", "error", err.Error())
		}

		return "", "", "", nil
	}
	return *result.Credentials.AccessKeyId, *result.Credentials.SecretAccessKey, *result.Credentials.SessionToken, nil
}

func CreateAwsSecretSession(provider string, region string, sessionName string, logger *zap.SugaredLogger) (awsSecSvc *secretsmanager.SecretsManager) {

	accessKey, secretID, sessiontoken, stserr := GetCredsFromSTS(logger)

	if stserr != nil {
		logger.Errorw("Error getting Credentials", "error", stserr)
	}

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretID, sessiontoken),
	})

	if err != nil {
		logger.Errorw("Can't start AWS session", "error", err)
	}

	awsSecSvc = secretsmanager.New(sess)

	return awsSecSvc
}

func CreateAwsSecret(clusterName string, clusterID string, token string, region string, logger *zap.SugaredLogger) (string, error) {
	sessionName := "AWS create sercet sesion, at " + time.Stamp
	awsSecSvc := CreateAwsSecretSession("aws", region, sessionName, logger)

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
	sessionName := "AWS Get sercet sesion, at " + time.Stamp
	awsSecSvc := CreateAwsSecretSession("aws", region, sessionName, logger)

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
