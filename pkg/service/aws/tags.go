package aws

import (
	"context"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/errors"
)

func asTags(labels map[string]string) []*ec2.Tag {
	tags := []*ec2.Tag{}
	for k, v := range labels {
		key := k
		val := v
		tags = append(tags, &ec2.Tag{
			Key:   &key,
			Value: &val,
		})
	}
	return tags
}

func (a *AWSController) addTag(ctx context.Context, region, clusterName, accountName string, labels map[string]string) error {
	session, err := NewSession(ctx, region, accountName)

	a.logger.Debugf("fetching cluster status for '%s', region '%s'", clusterName, region)
	if err != nil {
		return errors.Wrap(err, "addTag")
	}
	eksClient := session.getEksClient()

	cluster, err := getClusterSpec(ctx, eksClient, clusterName)
	if err != nil {
		return errors.Wrap(err, "addTag: ")
	}

	a.logger.Infow("tagging nodes", "cluster", cluster.Name)
	ec := session.getEC2Client()

	res, err := ec.DescribeInstances(&ec2.DescribeInstancesInput{})

	if err != nil {
		a.logger.Errorw("get instance", "error", err)
		return err
	}
	a.logger.Infow("ec2 nodes")

	if len(res.Reservations) == 0 {
		a.logger.Errorw("no ec2 instance found")
		return nil
	}

	rids := []*string{}
	for _, r := range res.Reservations[0].Instances {
		skip := false
		for _, t := range r.Tags {

			if *t.Key == "eks:cluster-name" && *t.Value != clusterName {
				skip = true
			}

			//Note : can pick nodes if we want node level granularity too

		}

		if skip {
			continue
		}
		id := *r.InstanceId
		rids = append(rids, &id)
	}

	if len(rids) == 0 {
		a.logger.Infow("no resources to tag")
		return errors.New("no instances in cluster to tag")
	}

	a.logger.Infow("adding tags to the following resources", "id", rids)
	tags := asTags(labels)
	_, err = ec.CreateTags(&ec2.CreateTagsInput{
		Resources: rids,
		Tags:      tags,
	})

	if err != nil {
		a.logger.Errorw("failed to add tags", "error", err)
		return err
	}
	return nil

}
