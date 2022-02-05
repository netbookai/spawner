package aws

import (
	"encoding/base64"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/route53"
	"gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice/system"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/aws-iam-authenticator/pkg/token"
)

type Session struct {
	AwsSession *session.Session
	Region     string
	TeamId     string
}

func NewSession(region string, accountName string) (*Session, error) {

	credentials, err := system.GetAwsCredentials(region, accountName)

	if err != nil {
		return nil, err
	}

	//get credentials for the user of given team id
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials,
	})

	if err != nil {
		return nil, err
	}

	return &Session{
		TeamId:     accountName,
		Region:     region,
		AwsSession: sess,
	}, nil
}

func newKubeConfig(session *session.Session, cluster *eks.Cluster) (*rest.Config, error) {
	gen, err := token.NewGenerator(true, false)
	if err != nil {
		return nil, err
	}
	opts := &token.GetTokenOptions{
		ClusterID: aws.StringValue(cluster.Name),
		Session:   session,
	}
	tok, err := gen.GetWithOptions(opts)
	if err != nil {
		return nil, err
	}
	ca, err := base64.StdEncoding.DecodeString(aws.StringValue(cluster.CertificateAuthority.Data))
	if err != nil {
		return nil, err
	}
	return &rest.Config{
		Host:        aws.StringValue(cluster.Endpoint),
		BearerToken: tok.Token,
		TLSClientConfig: rest.TLSClientConfig{
			CAData: ca,
		},
	}, nil
}

func newClientset(session *session.Session, cluster *eks.Cluster) (*kubernetes.Clientset, error) {
	config, err := newKubeConfig(session, cluster)
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

//---

func (ses *Session) getEksClient() *eks.EKS {
	return eks.New(ses.AwsSession)
}

func (ses *Session) getEC2Client() *ec2.EC2 {
	return ec2.New(ses.AwsSession)
}

func (ses *Session) getK8sClient(cluster *eks.Cluster) (*kubernetes.Clientset, error) {
	return newClientset(ses.AwsSession, cluster)
}

func (ses *Session) getIAMClient() *iam.IAM {
	return iam.New(ses.AwsSession)
}

func (ses *Session) getKubeConfig(cluster *eks.Cluster) (*rest.Config, error) {
	return newKubeConfig(ses.AwsSession, cluster)
}

func (ses *Session) getRoute53Client() *route53.Route53 {
	return route53.New(ses.AwsSession)
}
