package system

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/pkg/errors"
	"gitlab.com/netbook-devs/spawner-service/pkg/config"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/constants"
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
	conf := config.Get()

	var cred *credentials.Credentials

	if conf.Env == "local" {
		log.Println("running in dev mode, using ", conf.AWSAccessID)
		cred = credentials.NewStaticCredentials(conf.AWSAccessID, conf.AWSSecretKey, conf.AWSToken)

	} else {
		stsCreds, stserr := getSystemCredential()

		if stserr != nil {
			return nil, stserr
		}
		cred = credentials.NewStaticCredentials(*stsCreds.AccessKeyId, *stsCreds.SecretAccessKey, *stsCreds.SessionToken)
	}

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: cred,
	})

	return sess, err
}

func getSecretManager(region string) (*secretsmanager.SecretsManager, error) {

	sess, err := createSession(region)
	if err != nil {
		return nil, err
	}

	secretManager := secretsmanager.New(sess)

	return secretManager, nil
}

//GetAwsCredentials Retrieve user credentials from the secret manager
func GetAwsCredentials(ctx context.Context, region, accountName string) (*credentials.Credentials, error) {
	c, err := GetCredentials(ctx, region, accountName, constants.CredAws)
	if err != nil {
		return nil, err
	}
	return credentials.NewStaticCredentials(c.GetAws().Id, c.GetAws().Secret, c.GetAws().Token), nil
}

//sid create secrete id using credType and account name
func sid(credType, name string) string {
	return fmt.Sprintf("%s/%s", credType, name)
}

//GetCredentials retrieve crendential for the given cred type of a account
func GetCredentials(ctx context.Context, region, accountName, credType string) (Credentials, error) {
	secret, err := getSecretManager(region)
	if err != nil {
		return nil, errors.Wrapf(err, "GetCredentials: failed to get secretsmanager")
	}
	s := sid(credType, accountName)
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     &s,
		VersionStage: aws.String("AWSCURRENT"),
	}

	result, err := secret.GetSecretValueWithContext(ctx, input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			return nil, errors.Wrapf(aerr, "GetAwsCredentials: failed to fetch user credentials")
		}
	}

	var cred Credentials
	switch credType {
	case constants.CredAws:
		cred, err = NewAwsCredential(*result.SecretString)
	case constants.CredAzure:
		cred, err = NewAzureCredential(*result.SecretString)
<<<<<<< HEAD
	case constants.CredGitPat:
		cred, err = NewGitPAT(*result.SecretString)
||||||| parent of c438c6f (implement get clusters, cluster status, delete cluster, credentials for gcp)
=======
	case constants.GcpLabel:
		cred, err = NewGcpCredential(*result.SecretString)
>>>>>>> c438c6f (implement get clusters, cluster status, delete cluster, credentials for gcp)
	}

	if err != nil {
		return nil, errors.Wrap(err, "GetCredentials")
	}
	return cred, nil
}

//WriteOrUpdateCredential Creates a new secrets in AWS, updates the existing if key already present
// update will be set to true when key Update operation is perfromed,
// false on new secret creation
func WriteOrUpdateCredential(ctx context.Context, region, account, credType string, cred Credentials) (update bool, err error) {

	secret, err := getSecretManager(region)
	if err != nil {
		return false, err
	}
	value := cred.AsSecretValue()

	s := sid(credType, account)
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     &s,
		VersionStage: aws.String("AWSCURRENT"),
	}

	exist := true
	result, err := secret.GetSecretValueWithContext(ctx, input)
	if err != nil {
		//check if key exist
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == secretsmanager.ErrCodeResourceNotFoundException {
				exist = false
			}
		}
		//we will ignore any other might have accured, which is most lilkey to get caugh next,
		//handling error here becomes tedius,
	}

	if !exist {

		_, err = secret.CreateSecretWithContext(ctx, &secretsmanager.CreateSecretInput{
			Name:         &s,
			SecretString: &value,
		})
		update = false
	} else {
		_, err = secret.UpdateSecretWithContext(ctx, &secretsmanager.UpdateSecretInput{

			SecretId:     result.Name,
			SecretString: &value,
		})
		update = true
	}
	return

}
