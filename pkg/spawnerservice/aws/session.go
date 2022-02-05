package aws

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/pkg/errors"
)

type AWSStsCreds struct {
	accessKey       *string
	secretAccesskey *string
	sessionToken    *string
}

func GetCredsFromSTS() (AWSStsCreds, error) {
	ses, err := session.NewSession()
	if err != nil {
		return AWSStsCreds{}, err
	}

	svc := sts.New(ses)
	web_identity_token, err := os.ReadFile("/var/run/secrets/eks.amazonaws.com/serviceaccount/token")
	if err != nil {
		return AWSStsCreds{}, errors.Wrap(err, "error reading web identity token")
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

		return AWSStsCreds{}, intErr
	}
	return AWSStsCreds{
		accessKey:       result.Credentials.AccessKeyId,
		secretAccesskey: result.Credentials.SecretAccessKey,
		sessionToken:    result.Credentials.SessionToken,
	}, nil
}

func CreateBaseSession(region string) (*session.Session, error) {
	awsStsCreds, stserr := GetCredsFromSTS()

	if stserr != nil {
		return &session.Session{}, stserr
	}

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(*awsStsCreds.accessKey, *awsStsCreds.secretAccesskey, *awsStsCreds.sessionToken),
	})

	// For local testing
	//sess, err := session.NewSession(&aws.Config{
	//	Region:      aws.String(region),
	//	Credentials: credentials.NewStaticCredentials("", "", ""),
	//})
	//if err != nil {
	//	fmt.Println(" Failed to get session for local run", err)
	//}

	return sess, err
}
