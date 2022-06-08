package system

import (
	"context"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/google/uuid"
	"github.com/libdns/libdns"
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

func getHostedZoneFromZoneId(route53Client *route53.Route53, zoneId string) (string, error) {

	hostedZoneOutput, err := route53Client.GetHostedZone(&route53.GetHostedZoneInput{
		Id: &zoneId,
	})

	if err != nil {
		return "", errors.Wrap(err, "getHostedZoneFromZoneId: failed to get route53 hostedZone ")
	}

	return *hostedZoneOutput.HostedZone.Name, nil

}

func applyChange(ctx context.Context, route53Client *route53.Route53, input *route53.ChangeResourceRecordSetsInput) error {
	changeResult, err := route53Client.ChangeResourceRecordSetsWithContext(ctx, input)
	if err != nil {
		return err
	}

	changeInput := &route53.GetChangeInput{
		Id: changeResult.ChangeInfo.Id,
	}

	err = route53Client.WaitUntilResourceRecordSetsChangedWithContext(ctx, changeInput)
	if err != nil {
		return err
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
func GetRoute53TXTRecords(ctx context.Context, regionName string) ([]types.Route53Record, error) {

	hostedZoneId := config.Get().AwsRoute53HostedZoneID

	// Creating AWS Route53 session
	route53Client, err := getRoute53Sess(regionName)
	if err != nil {
		return nil, errors.Wrap(err, "GetRoute53Records: failed to create route53 session")
	}

	hostedZone, err := getHostedZoneFromZoneId(route53Client, hostedZoneId)

	if err != nil {
		return nil, errors.Wrap(err, "GetRoute53Records: failed to get route53 hostedZone ")
	}

	getRecordsInput := &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(hostedZoneId),
		MaxItems:     aws.String("1000"),
	}

	var records []types.Route53Record
	var recordSets []*route53.ResourceRecordSet

	for {
		getRecordResult, err := route53Client.ListResourceRecordSetsWithContext(ctx, getRecordsInput)
		if err != nil {
			return nil, errors.Wrap(err, "GetRoute53Records: failed to get route53 records")
		}

		recordSets = append(recordSets, getRecordResult.ResourceRecordSets...)
		if *getRecordResult.IsTruncated {
			getRecordsInput.StartRecordName = getRecordResult.NextRecordName
			getRecordsInput.StartRecordType = getRecordResult.NextRecordType
			getRecordsInput.StartRecordIdentifier = getRecordResult.NextRecordIdentifier
		} else {
			break
		}
	}

	for _, recordSet := range recordSets {
		// filtering txt records only
		if *recordSet.Type != route53.RRTypeTxt {
			continue
		}

		for _, resourceRecord := range recordSet.ResourceRecords {
			record := types.Route53Record{
				Name:         libdns.AbsoluteName(*recordSet.Name, hostedZone),
				Value:        *resourceRecord.Value,
				Type:         *recordSet.Type,
				TTLInSeconds: int64(*recordSet.TTL),
			}

			records = append(records, record)
		}
	}

	return records, nil

}

func deleteRoute53Record(ctx context.Context, route53Client *route53.Route53, record types.Route53Record, hostedZoneId, hostedZone string) error {

	deleteInput := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("DELETE"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(libdns.AbsoluteName(record.Name, hostedZone)),
						ResourceRecords: []*route53.ResourceRecord{
							{
								Value: aws.String(record.Value),
							},
						},
						TTL:  aws.Int64(int64(record.TTLInSeconds)),
						Type: aws.String(record.Type),
					},
				},
			},
		},
		HostedZoneId: aws.String(hostedZoneId),
	}

	err := applyChange(ctx, route53Client, deleteInput)

	if err != nil {
		return errors.Wrap(err, "deleteRoute53Record: failed to delete record")
	}

	return nil

}

// DeleteRoute53Records deletes the records from the zone. If a record does not have an ID,
// it will be looked up. It returns the records that were deleted.
func DeleteRoute53Records(ctx context.Context, regionName string, records []types.Route53Record) ([]types.Route53Record, error) {

	hostedZoneId := config.Get().AwsRoute53HostedZoneID

	// Creating AWS Route53 session
	route53Client, err := getRoute53Sess(regionName)
	if err != nil {
		return nil, errors.Wrap(err, "DeleteRoute53Records: failed to create route53 session")
	}

	hostedZone, err := getHostedZoneFromZoneId(route53Client, hostedZoneId)

	if err != nil {
		return nil, errors.Wrap(err, "DeleteRoute53Records: failed to get route53 hostedZone ")
	}

	var deletedRecords []types.Route53Record

	for _, record := range records {
		err := deleteRoute53Record(ctx, route53Client, record, hostedZoneId, hostedZone)
		if err != nil {
			return deletedRecords, err
		}
		deletedRecords = append(deletedRecords, record)
	}

	return deletedRecords, nil
}

func createRoute53Record(ctx context.Context, route53Client *route53.Route53, zoneID string, record types.Route53Record, zone string) (types.Route53Record, error) {
	// AWS Route53 TXT record value must be enclosed in quotation marks on create
	if record.Type == route53.RRTypeTxt {
		record.Value = strconv.Quote(record.Value)
	}

	createInput := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("CREATE"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(libdns.AbsoluteName(record.Name, zone)),
						ResourceRecords: []*route53.ResourceRecord{
							{
								Value: aws.String(record.Value),
							},
						},
						TTL:  aws.Int64(int64(record.TTLInSeconds)),
						Type: aws.String(record.Type),
					},
				},
			},
		},
		HostedZoneId: aws.String(zoneID),
	}

	err := applyChange(ctx, route53Client, createInput)
	if err != nil {
		return record, err
	}

	return record, nil
}

// AppendRoute53Records adds records to the zone. It returns the records that were added.
func AppendRoute53Records(ctx context.Context, regionName string, records []types.Route53Record) ([]types.Route53Record, error) {

	hostedZoneId := config.Get().AwsRoute53HostedZoneID

	// Creating AWS Route53 session
	route53Client, err := getRoute53Sess(regionName)
	if err != nil {
		return nil, errors.Wrap(err, "AddRoute53Record: failed to create route53 session")
	}

	hostedZone, err := getHostedZoneFromZoneId(route53Client, hostedZoneId)

	if err != nil {
		return nil, errors.Wrap(err, "AddRoute53Record: failed to get route53 hostedZone ")
	}

	var createdRecords []types.Route53Record

	for _, record := range records {
		newRecord, err := createRoute53Record(ctx, route53Client, hostedZoneId, record, hostedZone)
		if err != nil {
			return createdRecords, err
		}
		createdRecords = append(createdRecords, newRecord)
	}

	return createdRecords, nil
}
