package aws

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/google/uuid"
	"gitlab.com/netbook-devs/spawner-service/pb"
)

func CreateAwsRoute53Session(region string) (*route53.Route53, error) {
	sess, err := CreateBaseSession(region)
	if err != nil {
		return nil, err
	}

	awsAwsRoute53Sess := route53.New(sess)

	return awsAwsRoute53Sess, nil
}

func (svc AWSController) AddRoute53Record(ctx context.Context, req *pb.AddRoute53RecordRequest) (*pb.AddRoute53RecordResponse, error) {
	logger := svc.logger

	regionClassicLoadBalancerHostedID := map[string]string{
		"us-east-2":      "Z3AADJGX6KTTL2",
		"us-east-1":      "Z35SXDOTRQ7X7K",
		"us-west-1":      "Z368ELLRRE2KJ0",
		"us-west-2":      "Z1H1FL5HABSF5",
		"af-south-1":     "Z268VQBMOI5EKX",
		"ap-east-1":      "Z3DQVH9N71FHZ0",
		"ap-southeast-3": "Z08888821HLRG5A9ZRTER",
		"ap-south-1":     "ZP97RAFLXTNZK",
		"ap-northeast-3": "Z5LXEXXYW11ES",
		"ap-northeast-2": "ZWKZPGTI48KDX",
		"ap-southeast-1": "Z1LMS91P8CMLE5",
		"ap-southeast-2": "Z1GM3OXH4ZPM65",
		"ap-northeast-1": "Z14GRHDCWA56QT",
		"ca-central-1":   "ZQSVJUPU6J1EY",
		"cn-north-1":     "Z1GDH35T77C1KE",
		"cn-northwest-1": "ZM7IZAIOVVDZF",
		"eu-central-1":   "Z215JYRZR1TBD5",
		"eu-west-1":      "Z32O12XQLNTSW2",
		"eu-west-2":      "ZHURV8PSTC4K8",
		"eu-south-1":     "Z3ULH7SSC9OV64",
		"eu-west-3":      "Z3Q77PNBQS71R4",
		"eu-north-1":     "Z23TAZ6LKFMNIO",
		"me-south-1":     "ZS929ML54UICD",
		"sa-east-1":      "Z2P70J7HTTTPLU",
		"us-gov-east-1":  "Z166TLBEWOO7G0",
		"us-gov-west-1":  "Z33AYJ8TM3BH4J",
	}
	dnsName := req.GetDnsName()
	recordName := req.GetRecordName()
	regionName := req.GetRegionName()
	id, iderr := uuid.NewRandom()

	if iderr != nil {
		LogError("Failed creating random UUID", logger, iderr)
		return &pb.AddRoute53RecordResponse{}, iderr
	}

	_, ok := regionClassicLoadBalancerHostedID[regionName]
	if !ok {
		logger.Errorw("Region does not have matching ELB HostedZoneId")
		res := &pb.AddRoute53RecordResponse{
			Status: "Failed",
			Error:  "",
		}
		return res, errors.New("Region does not have matching ELB HostedZoneId")
	}

	hostedZoenId := svc.config.AwsRoute53HostedZoneId

	input := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("CREATE"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						AliasTarget: &route53.AliasTarget{
							// This DNS name is Classic Load Balancer URL
							DNSName:              aws.String("dualstack." + dnsName),
							EvaluateTargetHealth: aws.Bool(true),
							HostedZoneId:         aws.String(regionClassicLoadBalancerHostedID[regionName]),
						},
						Name:          aws.String(recordName),
						Region:        aws.String(regionName),
						SetIdentifier: aws.String(id.String()),
						Type:          aws.String("A"),
					},
				},
			},
			Comment: aws.String("ELB load balancers for " + recordName),
		},
		HostedZoneId: aws.String(hostedZoenId), //Z08929991BJOO3WMC7L0Q
	}
	// Creating AWS Route53 session
	awsAwsRoute53Sess, err := CreateAwsRoute53Session(regionName)

	if err != nil {
		logger.Errorw("Can't start AWS session", "error", err)
		return nil, err
	}

	result, err := awsAwsRoute53Sess.ChangeResourceRecordSets(input)

	if err != nil {
		LogError("AddAwsRoute53Record", logger, err)
		return &pb.AddRoute53RecordResponse{}, err
	}

	err = awsAwsRoute53Sess.WaitUntilResourceRecordSetsChanged(&route53.GetChangeInput{
		Id: *&result.ChangeInfo.Id,
	})

	if err != nil {
		LogError("WaitAddAwsRoute53Record", logger, err)
		return &pb.AddRoute53RecordResponse{}, err
	}

	res := &pb.AddRoute53RecordResponse{
		Status: *result.ChangeInfo.Id,
	}

	return res, nil
}
