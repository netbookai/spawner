package system

import (
	"context"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"gitlab.com/netbook-devs/spawner-service/pkg/config"
	"gitlab.com/netbook-devs/spawner-service/pkg/types"
)

var regionClassicLoadBalancerHostedID = map[string]string{
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

const emptyRegion = ""

func getLbHosterID(regionName string) (string, error) {
	id, ok := regionClassicLoadBalancerHostedID[regionName]
	if !ok {
		return "", errors.Errorf("region '%s' does not have matching ELB HostedZoneId", regionName)
	}

	return id, nil
}

func getRoute53Sess(region string) (*route53.Route53, error) {
	sess, err := createSession(region)
	if err != nil {
		return nil, err
	}

	route53Sess := route53.New(sess)
	return route53Sess, nil
}

func applyChange(ctx context.Context, records []types.Route53ResourceRecordSet, action string) error {

	hostedZoneId := config.Get().AwsRoute53HostedZoneID

	// Creating AWS Route53 session
	route53Client, err := getRoute53Sess(emptyRegion)
	if err != nil {
		return errors.Wrap(err, "applyChange: failed to create route53 session")
	}

	input := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: make([]*route53.Change, 0, len(records)),
		},
		HostedZoneId: aws.String(hostedZoneId),
	}

	for _, record := range records {

		changeRequest := &route53.Change{
			Action: aws.String(action),
			ResourceRecordSet: &route53.ResourceRecordSet{
				Name:            aws.String(record.Name),
				ResourceRecords: make([]*route53.ResourceRecord, 0, len(record.ResourceRecords)),
				TTL:             aws.Int64(int64(record.TTLInSeconds)),
				Type:            aws.String(record.Type),
			},
		}

		for _, resourceRecord := range record.ResourceRecords {

			// AWS Route53 TXT record value must be enclosed in quotation marks on create
			if record.Type == route53.RRTypeTxt {
				resourceRecord.Value = strconv.Quote(resourceRecord.Value)
			}

			changeRequest.ResourceRecordSet.ResourceRecords = append(changeRequest.ResourceRecordSet.ResourceRecords, &route53.ResourceRecord{
				Value: aws.String(resourceRecord.Value),
			})

		}

		input.ChangeBatch.Changes = append(input.ChangeBatch.Changes, changeRequest)

	}

	changeResult, err := route53Client.ChangeResourceRecordSetsWithContext(ctx, input)
	if err != nil {
		return errors.Wrap(err, "applyChange")
	}

	changeInput := &route53.GetChangeInput{
		Id: changeResult.ChangeInfo.Id,
	}

	err = route53Client.WaitUntilResourceRecordSetsChangedWithContext(ctx, changeInput)
	if err != nil {
		return errors.Wrap(err, "applyChange")
	}

	return nil
}

func AddRoute53Record(ctx context.Context, dnsName, recordName, regionName string, isAwsResource bool) (string, error) {
	id, err := uuid.NewRandom()

	if err != nil {
		return "", errors.Wrap(err, "AddRoute53Record: failed to get uuid ")
	}

	hostedZoenId := config.Get().AwsRoute53HostedZoneID

	input := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("CREATE"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name:          aws.String(recordName),
						SetIdentifier: aws.String(id.String()),
						Type:          aws.String("A"),
					},
				},
			},
			Comment: aws.String("ELB load balancers for " + recordName),
		},
		HostedZoneId: aws.String(hostedZoenId), //Z08929991BJOO3WMC7L0Q
	}

	if isAwsResource {
		// This is currently only for ELB urls
		classicLbHostedID, err := getLbHosterID(regionName)
		if err != nil {
			return "", errors.Wrap(err, "AddRoute53Record")
		}
		input.ChangeBatch.Changes[0].ResourceRecordSet.Region = aws.String(regionName)
		input.ChangeBatch.Changes[0].ResourceRecordSet.AliasTarget = &route53.AliasTarget{
			// This DNS name is Classic Load Balancer URL
			DNSName:              aws.String("dualstack." + dnsName),
			EvaluateTargetHealth: aws.Bool(true),
			HostedZoneId:         aws.String(classicLbHostedID),
		}
	} else {
		input.ChangeBatch.Changes[0].ResourceRecordSet.ResourceRecords = []*route53.ResourceRecord{{
			Value: aws.String(dnsName),
		}}
		input.ChangeBatch.Changes[0].ResourceRecordSet.TTL = aws.Int64(60)
		input.ChangeBatch.Changes[0].ResourceRecordSet.Weight = aws.Int64(1)
	}

	// Creating AWS Route53 session
	route53Client, err := getRoute53Sess(regionName)
	if err != nil {
		return "", errors.Wrap(err, "AddRoute53Record: failed to create route53 session")
	}

	result, err := route53Client.ChangeResourceRecordSets(input)

	if err != nil {
		return "", errors.Wrap(err, "AddAwsRoute53Record: ChangeResourceRecordSets returned error")
	}

	// logger.Infow("added route53 record set " + recordName)
	err = route53Client.WaitUntilResourceRecordSetsChanged(&route53.GetChangeInput{
		Id: result.ChangeInfo.Id,
	})

	if err != nil {
		return "", errors.Wrap(err, "AddAwsRoute53Record: WaitUntilResourceRecordSetsChanged returned error")
	}

	return *result.ChangeInfo.Id, nil
}

// GetRoute53TXTRecords will return TXT records only
func GetRoute53TXTRecords(ctx context.Context) ([]types.Route53ResourceRecordSet, error) {

	hostedZoneId := config.Get().AwsRoute53HostedZoneID

	// Creating AWS Route53 session
	route53Client, err := getRoute53Sess(emptyRegion)
	if err != nil {
		return nil, errors.Wrap(err, "GetRoute53Records: failed to create route53 session")
	}

	getRecordsInput := &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(hostedZoneId),
		MaxItems:     aws.String("1000"),
	}

	var resourceRecordSets []types.Route53ResourceRecordSet
	var r53recordSets []*route53.ResourceRecordSet

	// running this loop until we get all the records, getRecordResult.IsTruncated will be false
	// when we get the last records, route53 returns records in sets of size we mention
	// in this case we have mentioned value 1000 in input
	for {
		getRecordResult, err := route53Client.ListResourceRecordSetsWithContext(ctx, getRecordsInput)
		if err != nil {
			return nil, errors.Wrap(err, "GetRoute53Records: failed to get route53 records")
		}

		r53recordSets = append(r53recordSets, getRecordResult.ResourceRecordSets...)
		if *getRecordResult.IsTruncated {
			getRecordsInput.StartRecordName = getRecordResult.NextRecordName
			getRecordsInput.StartRecordType = getRecordResult.NextRecordType
			getRecordsInput.StartRecordIdentifier = getRecordResult.NextRecordIdentifier
		} else {
			break
		}
	}

	for _, recordSet := range r53recordSets {
		// filtering txt records only
		if *recordSet.Type != route53.RRTypeTxt {
			continue
		}

		resourceRecordSet := types.Route53ResourceRecordSet{
			Name:            *recordSet.Name,
			ResourceRecords: make([]types.ResourceRecordValue, 0, len(recordSet.ResourceRecords)),
			Type:            *recordSet.Type,
			TTLInSeconds:    int64(*recordSet.TTL),
		}

		for _, rr := range recordSet.ResourceRecords {

			resourceRecordSet.ResourceRecords = append(resourceRecordSet.ResourceRecords, types.ResourceRecordValue{
				Value: *rr.Value,
			})
		}

		resourceRecordSets = append(resourceRecordSets, resourceRecordSet)
	}

	return resourceRecordSets, nil

}

// CreateRoute53Records creates new records in the zone
func CreateRoute53Records(ctx context.Context, records []types.Route53ResourceRecordSet) error {

	err := applyChange(ctx, records, route53.ChangeActionCreate)

	if err != nil {
		return errors.Wrap(err, "CreateRoute53Records: failed to create record")
	}

	return nil
}

// DeleteRoute53Records deletes the records from the zone
func DeleteRoute53Records(ctx context.Context, records []types.Route53ResourceRecordSet) error {

	err := applyChange(ctx, records, route53.ChangeActionDelete)

	if err != nil {
		return errors.Wrap(err, "DeleteRoute53Records: failed to delete records")
	}

	return nil
}
