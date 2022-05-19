package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
)

func (a *AWSController) createSpotLaunchTemplate(ctx context.Context, client *ec2.EC2, amiId, price, instance, volumeType string, diskSize int64, label map[string]string) (string, error) {

	//amiId = "ami-09dd25549dd970de5"
	//fetch launch template ]
	//	marketTyp := "spot"
	//	spotType := "one-time" // "persistent" -- cant have persistent for autoscaling instacnes
	//	// https://aws.amazon.com/blogs/aws/new-ec2-spot-blocks-for-defined-duration-workloads/
	//	blockDuration := int64(360) //6hours
	//	interruptionBehaviour := "terminate"

	tags := asTags(label)
	ltd := ec2.RequestLaunchTemplateData{
		ImageId: &amiId,
		//	InstanceMarketOptions: &ec2.LaunchTemplateInstanceMarketOptionsRequest{
		//		MarketType: &marketTyp,
		//		SpotOptions: &ec2.LaunchTemplateSpotMarketOptionsRequest{
		//			BlockDurationMinutes:         &blockDuration,
		//			InstanceInterruptionBehavior: &interruptionBehaviour,
		//			MaxPrice:                     &price,
		//			SpotInstanceType:             &spotType,
		//		},
		//	},
		InstanceType: &instance,
		BlockDeviceMappings: []*ec2.LaunchTemplateBlockDeviceMappingRequest{
			{
				DeviceName: aws.String("/dev/xvda"),
				Ebs: &ec2.LaunchTemplateEbsBlockDeviceRequest{
					DeleteOnTermination: aws.Bool(true),
					VolumeSize:          &diskSize,
					VolumeType:          &volumeType,
				},
			}},

		TagSpecifications: []*ec2.LaunchTemplateTagSpecificationRequest{
			{
				ResourceType: aws.String("instance"),
				Tags:         tags,
			},
			{
				ResourceType: aws.String("volume"),
				Tags:         tags,
			},
			{
				ResourceType: aws.String("elastic-gpu"),
				Tags:         tags,
			},
			//			{
			//				ResourceType: aws.String("spot-instance-request"),
			//				Tags:         tags,
			//			},
			{
				ResourceType: aws.String("network-interface"),
				Tags:         tags,
			},
		},
	}

	s := time.Now().Second()
	name := fmt.Sprintf("spt-lt-%d", s)
	fmt.Println("creating tenplate  ", name)
	res, err := client.CreateLaunchTemplateWithContext(ctx, &ec2.CreateLaunchTemplateInput{

		LaunchTemplateData: &ltd,
		LaunchTemplateName: aws.String(name),
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("launch-template"),
				Tags:         tags,
			},
			//			{
			//				ResourceType: aws.String("volume"),
			//				Tags:         tags,
			//			},
			//			{
			//				ResourceType: aws.String("elastic-gpu"),
			//				Tags:         tags,
			//			},
			//			{
			//				ResourceType: aws.String("spot-instance-request"),
			//				Tags:         tags,
			//			},
			//			{
			//				ResourceType: aws.String("network-interface"),
			//				Tags:         tags,
			//			},
		},
	})

	if err != nil {
		return "", errors.Wrap(err, "createSpotLaunchTemplate")
	}

	l, err := client.DescribeLaunchTemplateVersions(&ec2.DescribeLaunchTemplateVersionsInput{
		LaunchTemplateId: res.LaunchTemplate.LaunchTemplateId,
		MaxVersion:       aws.String("3"),
		MinVersion:       aws.String("0"),
	})

	if err != nil {
		a.logger.Errorw("failed to get the launch template version", "error", err)
	} else {
		spew.Dump(l.LaunchTemplateVersions)
	}

	return *res.LaunchTemplate.LaunchTemplateId, nil
}
