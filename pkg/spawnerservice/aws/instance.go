package aws

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice/constants"
)

func WaitTillInstanceTerminated(region string, instanceLabelMap map[string]string) error {
	sess, err := Ec2SessionFactory(region)
	if err != nil {
		return err
	}

	return sess.WaitUntilInstanceTerminated(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String(fmt.Sprintf("tag:%s", constants.NODE_NAME_LABEL)),
				Values: aws.StringSlice([]string{instanceLabelMap[constants.NODE_NAME_LABEL]}),
			},
		},
	})
}

// Gets all regions, AZs and instances avilable in each AZ
// Writes the following maps
// ---- regionAzMap.json : map of region -> list of AZs in region
// ---- azInstanceMap.json : map of AZ -> list of inststances in AZ
// ---- instanceSpecCache.json : map of instance -> details of instance
func GetAwsInstances() error {
	regionAzMap := map[string][]string{}
	azInstanceMap := map[string][]string{}
	instanceSpecCache := map[string]*ec2.InstanceTypeInfo{}

	sess, err := Ec2SessionFactory("us-west-2")
	if err != nil {
		fmt.Println("error creating session")
		return err
	}

	regions, err := sess.DescribeRegions(&ec2.DescribeRegionsInput{
		AllRegions: aws.Bool(true),
	})
	if err != nil {
		fmt.Println("error getting regions")
		return err
	}

	for _, region := range regions.Regions {
		sess, err := Ec2SessionFactory(*region.RegionName)
		if err != nil {
			fmt.Println("error creating session for region " + *region.RegionName)
			return err
		}

		regionAzMap[*region.RegionName] = []string{}

		azsInRegion, err := sess.DescribeAvailabilityZones(&ec2.DescribeAvailabilityZonesInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("region-name"),
					Values: []*string{region.RegionName},
				},
			},
			AllAvailabilityZones: aws.Bool(true),
		})
		if err != nil {
			//AuthFailure
			if aerr, ok := err.(awserr.Error); ok {
				if aerr.Code() == "AuthFailure" {
					fmt.Println("auth error for region " + *region.RegionName)
					continue
				}
			}
			fmt.Println("error getting azs for region " + *region.RegionName)
			return err
		}
		fmt.Println("got azs for region " + *region.RegionName)

		for _, az := range azsInRegion.AvailabilityZones {
			regionAzMap[*region.RegionName] = append(regionAzMap[*region.RegionName], *az.ZoneName)
			azInstanceMap[*az.ZoneName] = []string{}

			instancesInAz, err := sess.DescribeInstanceTypeOfferings(&ec2.DescribeInstanceTypeOfferingsInput{
				Filters: []*ec2.Filter{
					{
						Name:   aws.String("location"),
						Values: []*string{az.ZoneName},
					},
				},
				LocationType: aws.String("availability-zone"),
			})
			if err != nil {
				fmt.Println("error getting instances for az " + *az.ZoneName)
				return err
			}

			for _, inst := range instancesInAz.InstanceTypeOfferings {
				azInstanceMap[*az.ZoneName] = append(azInstanceMap[*az.ZoneName], *inst.InstanceType)

				if _, ok := instanceSpecCache[*inst.InstanceType]; !ok {
					instance, err := sess.DescribeInstanceTypes(&ec2.DescribeInstanceTypesInput{
						InstanceTypes: []*string{
							inst.InstanceType,
						},
					})
					if err != nil {
						fmt.Println("error getting instance details for instance " + *inst.InstanceType)
						return err
					}

					instanceSpecCache[*inst.InstanceType] = instance.InstanceTypes[0]
				}
			}
		}
	}

	// readArr, err := os.ReadFile("test1.json")
	// if err != nil {
	// 	fmt.Println("error reading file")
	// 	return err
	// }

	// var insts []ec2.InstanceTypeOffering

	// err = json.Unmarshal(readArr, &insts)
	// if err != nil {
	// 	fmt.Println("error unmarshalling json")
	// 	return err
	// }

	// fmt.Println(insts)

	err = writeFile("regionAzMap.json", regionAzMap)
	if err != nil {
		fmt.Println("error writing regionAzMap json")
		return err
	}

	err = writeFile("azInstanceMap.json", azInstanceMap)
	if err != nil {
		fmt.Println("error writing azInstanceMap json")
		return err
	}

	err = writeFile("instanceSpecCache.json", instanceSpecCache)
	if err != nil {
		fmt.Println("error writing instanceSpecCache json")
		return err
	}

	return nil
}

// Writes data to filename in the current location
func writeFile(filename string, data interface{}) error {
	jsBytArr, err := json.Marshal(data)
	if err != nil {
		fmt.Println("error marshalling to json " + filename)
		return err
	}

	err = os.WriteFile(filename, jsBytArr, 0644)
	if err != nil {
		fmt.Println("error writing file")
		return err
	}

	return nil
}
