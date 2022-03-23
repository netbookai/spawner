package system

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/google/uuid"

	"gitlab.com/netbook-devs/spawner-service/pkg/config"
)

func getRoute53Sess(region string) (*route53.Route53, error) {
	sess, err := createSession(region)
	if err != nil {
		return nil, err
	}

	route53Sess := route53.New(sess)

	return route53Sess, nil
}

func AddRoute53Record(ctx context.Context, dnsName, recordName, regionName string) (string, error) {
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

	id, iderr := uuid.NewRandom()

	if iderr != nil {
		// LogError("Failed creating random UUID", logger, iderr)
		return "", iderr
	}

	_, ok := regionClassicLoadBalancerHostedID[regionName]
	if !ok {
		// logger.Errorw("Region does not have matching ELB HostedZoneId")
		return "", errors.New("Region does not have matching ELB HostedZoneId")
	}

	hostedZoenId := config.Get().AwsRoute53HostedZoneID

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
	route53Client, err := getRoute53Sess(regionName)
	if err != nil {
		return "", err
	}

	if err != nil {
		// logger.Errorw("Can't start AWS session", "error", err)
		return "", err
	}

	result, err := route53Client.ChangeResourceRecordSets(input)

	if err != nil {
		// LogError("AddAwsRoute53Record", logger, err)
		return "", err
	}

	// logger.Infow("added route53 record set " + recordName)

	err = route53Client.WaitUntilResourceRecordSetsChanged(&route53.GetChangeInput{
		Id: *&result.ChangeInfo.Id,
	})

	if err != nil {
		// LogError("WaitAddAwsRoute53Record", logger, err)
		return "", err
	}

	return *result.ChangeInfo.Id, nil
}
