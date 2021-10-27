package aws

import (
	"encoding/base64"
	"fmt"
	"log"
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

func GetCredsFromSTS() (string, string, string, error) {
	svc := sts.New(session.New())
	web_identity_token, err := os.ReadFile("/var/run/secrets/eks.amazonaws.com/serviceaccount/token")
	if err != nil {
		fmt.Errorf("Error reading token")
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
			case sts.ErrCodeMalformedPolicyDocumentException:
				fmt.Println(sts.ErrCodeMalformedPolicyDocumentException, aerr.Error())
			case sts.ErrCodePackedPolicyTooLargeException:
				fmt.Println(sts.ErrCodePackedPolicyTooLargeException, aerr.Error())
			case sts.ErrCodeIDPRejectedClaimException:
				fmt.Println(sts.ErrCodeIDPRejectedClaimException, aerr.Error())
			case sts.ErrCodeIDPCommunicationErrorException:
				fmt.Println(sts.ErrCodeIDPCommunicationErrorException, aerr.Error())
			case sts.ErrCodeInvalidIdentityTokenException:
				fmt.Println(sts.ErrCodeInvalidIdentityTokenException, aerr.Error())
			case sts.ErrCodeExpiredTokenException:
				fmt.Println(sts.ErrCodeExpiredTokenException, aerr.Error())
			case sts.ErrCodeRegionDisabledException:
				fmt.Println(sts.ErrCodeRegionDisabledException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}

	}
	return *result.Credentials.AccessKeyId, *result.Credentials.SecretAccessKey, *result.Credentials.SessionToken, nil
}

func CreateAwsSecretSession(provider string, region string, sessionName string, logger *zap.SugaredLogger) (awsSecSvc *secretsmanager.SecretsManager) {

	accessKey, secretID, sessiontoken, stserr := GetCredsFromSTS()
	if stserr != nil {
		log.Fatalf("Error getting Credentials")
	}

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretID, sessiontoken),
	})
	awsSecSvc = secretsmanager.New(sess)

	if err != nil {
		logger.Errorw("Can't start AWS session", "error", err)
	}
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
