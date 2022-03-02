package system

import (
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/pkg/errors"
)

//manages system level secrets,

//SystemCreds this is spawner service credentials
type SystemCreds struct {
	accessKey    *string
	secretKey    *string
	sessionToken *string
}

func getSystemCredential() (*sts.Credentials, error) {
	ses, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	svc := sts.New(ses)
	web_identity_token, err := os.ReadFile("/var/run/secrets/eks.amazonaws.com/serviceaccount/token")
	if err != nil {
		return nil, errors.Wrap(err, "error reading web identity token")
	}

	input := &sts.AssumeRoleWithWebIdentityInput{
		DurationSeconds:  aws.Int64(900),
		RoleArn:          aws.String(os.Getenv("AWS_ROLE_ARN")),
		RoleSessionName:  aws.String("SecretsConnection"),
		WebIdentityToken: aws.String(string(web_identity_token)),
	}

	result, err := svc.AssumeRoleWithWebIdentity(input)

	if err != nil {
		var intErr error
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case sts.ErrCodeExpiredTokenException:
				intErr = errors.Wrap(aerr, "token expired")
			case sts.ErrCodeRegionDisabledException:
				intErr = errors.Wrap(aerr, "error creating token: region disabled")
			default:
				intErr = errors.Wrap(aerr, "error accessing aws")
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			intErr = errors.Wrap(err, "error while getting credentials")
		}

		return nil, intErr
	}
	return result.Credentials, nil
}

//createSession create new application session
func createSession(region string) (*session.Session, error) {
	stsCreds, stserr := getSystemCredential()

	if stserr != nil {
		return nil, stserr
	}

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(*stsCreds.AccessKeyId, *stsCreds.SecretAccessKey, *stsCreds.SessionToken),
	})

	return sess, err
}

func parseCredentials(blob string) (*credentials.Credentials, error) {

	//secret_id,secret_key
	splits := strings.Split(blob, ",")
	if len(splits) != 2 {
		return nil, errors.New("invalid credentials found in secrets")
	}
	// token is set to blank for now
	creds := credentials.NewStaticCredentials(splits[0], splits[1], "")

	return creds, nil
}

//GetAwsCredentials Retrieve user credentials from the secret manager
func GetAwsCredentials(region, accountName string) (*credentials.Credentials, error) {
	sess, err := createSession(region)
	if err != nil {
		return nil, err
	}

	awsSecSvc := secretsmanager.New(sess)

	if err != nil {
		return nil, err
	}

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     &accountName,
		VersionStage: aws.String("AWSCURRENT"),
	}

	result, err := awsSecSvc.GetSecretValue(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			return nil, aerr
		}
	}
	return parseCredentials(*result.SecretString)
}
