package aws

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/errors"
	"gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice/constants"
)

type AwsWkspRegionNetworkStack struct {
	Vpc         *ec2.Vpc
	Gateway     *ec2.InternetGateway
	RouteTables []*ec2.RouteTable
	Subnets     []*ec2.Subnet
}

const (
	vpcCidr           = "192.168.0.0/16"
	vpcNameFmt        = "nb-wkps-vpc-%s"
	gatewayNameFmt    = "nb-wksp-internet-gateway-%s"
	routeTableNameFmt = "nb-wkps-route-table-%s"
	routeNameFmt      = "nb-wkps-route-%s"
	subnetNameFmt     = "nb-wksp-subnet-%s-%s"
	subnet01Cidr      = "192.168.64.0/18"
	subnet02Cidr      = "192.168.128.0/18"
	subnet03Cidr      = "192.168.192.0/18"
)

var (
	subnetUpto4Cidr = [4]string{"192.168.0.0/18", "192.168.64.0/18", "192.168.128.0/18", "192.168.192.0/18"}
	subnetUpto8Cidr = [8]string{"192.168.0.0/19", "192.168.32.0/19", "192.168.64.0/19", "192.168.96.0/19", "192.168.128.0/19", "192.168.160.0/19", "192.168.192.0/19", "192.168.224.0/19"}
)

func CreateAwsEc2Session(region string) (*ec2.EC2, error) {
	sess, err := CreateBaseSession(region)
	if err != nil {
		return nil, err
	}

	awsEc2Sess := ec2.New(sess)

	return awsEc2Sess, nil
}

func GetRegionWkspNetworkStack(region string) (*AwsWkspRegionNetworkStack, error) {
	vpcName := fmt.Sprintf(vpcNameFmt, region)

	rv := &AwsWkspRegionNetworkStack{}

	sess, err := CreateAwsEc2Session(region)
	if err != nil {
		return rv, errors.Wrapf(err, "error creating ec2 session for region %s", region)
	}

	vpcOut, err := sess.DescribeVpcs(&ec2.DescribeVpcsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String(fmt.Sprintf("tag:%s", constants.NAME_LABEL)),
				Values: aws.StringSlice([]string{vpcName}),
			},
			{
				Name:   aws.String(fmt.Sprintf("tag:%s", constants.CREATOR_LABEL)),
				Values: aws.StringSlice([]string{constants.SPAWNER_SERVICE_LABEL}),
			},
			{
				Name:   aws.String(fmt.Sprintf("tag:%s", constants.PROVISIONER_LABEL)),
				Values: aws.StringSlice([]string{constants.RANCHER_LABEL}),
			},
			{
				Name:   aws.String(fmt.Sprintf("tag:%s", constants.NB_TYPE_TAG_KEY)),
				Values: aws.StringSlice([]string{constants.NB_REGION_WKSP_NETWORK_STK}),
			},
		},
	})
	if err != nil {
		return rv, errors.Wrapf(err, "error getting vpcs from aws with name %s", vpcName)
	}

	var vpc *ec2.Vpc
	if len(vpcOut.Vpcs) > 0 {
		vpc = vpcOut.Vpcs[0]
	} else {
		return rv, nil
	}
	rv.Vpc = vpc

	igwOut, err := sess.DescribeInternetGateways(&ec2.DescribeInternetGatewaysInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("attachment.vpc-id"),
				Values: aws.StringSlice([]string{*vpc.VpcId}),
			},
		},
	})
	if err != nil {
		return rv, errors.Wrapf(err, "error getting internet gateway from aws attached to vpc %s", *vpc.VpcId)
	}
	if len(igwOut.InternetGateways) > 0 {
		rv.Gateway = igwOut.InternetGateways[0]
	}

	routeTblOut, err := sess.DescribeRouteTables(&ec2.DescribeRouteTablesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: aws.StringSlice([]string{*vpc.VpcId}),
			},
		},
	})
	if err != nil {
		return rv, errors.Wrapf(err, "error getting route table from aws attached to vpc %s", *vpc.VpcId)
	}
	if len(routeTblOut.RouteTables) > 0 {
		rv.RouteTables = routeTblOut.RouteTables
	}

	subnetOut, err := sess.DescribeSubnets(&ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String(fmt.Sprintf("tag:%s", constants.VPC_TAG_KEY)),
				Values: aws.StringSlice([]string{vpcName}),
			},
		},
	})
	if err != nil {
		return rv, errors.Wrapf(err, "error getting subnets from aws for vpc with name %s", *vpc.VpcId)
	}

	if len(subnetOut.Subnets) > 0 {
		rv.Subnets = subnetOut.Subnets
	} else {
		return rv, fmt.Errorf("subnets not associated with vpc %s", *vpc.VpcId)
	}

	return rv, nil
}

func DeleteRegionWkspNetworkStack(region string, netStk AwsWkspRegionNetworkStack) error {
	sess, err := CreateAwsEc2Session(region)
	if err != nil {
		return errors.Wrapf(err, "error creating ec2 session for region %s", region)
	}

	for _, subn := range netStk.Subnets {
		_, err := sess.DeleteSubnet(&ec2.DeleteSubnetInput{
			SubnetId: subn.SubnetId,
		})
		if err != nil {
			return errors.Wrapf(err, "error deleting subnet %s of vpc %s in region %s", *subn.SubnetId, *netStk.Vpc.VpcId, region)
		}
	}

	// RouteTables exist as 2 even when we create one
	// One as main and other as non-main
	// We can only delete non-main, main one is deleted when deleting vpc
	// Sorting Route tables by non-Main first followed by Main ones
	sort.Slice(netStk.RouteTables, func(i, j int) bool {
		return netStk.RouteTables[j].Associations != nil
	})
	for _, routeTbl := range netStk.RouteTables {
		if routeTbl.Associations == nil || len(routeTbl.Associations) == 0 || !*routeTbl.Associations[0].Main {
			_, err = sess.DeleteRouteTable(&ec2.DeleteRouteTableInput{
				RouteTableId: routeTbl.RouteTableId,
			})
			if err != nil {
				return errors.Wrapf(err, "error deleting route table %s in vpc %s in region %s", *routeTbl.RouteTableId, *netStk.Vpc.VpcId, region)
			}
		}
	}

	if netStk.Gateway != nil {
		_, err = sess.DetachInternetGateway(&ec2.DetachInternetGatewayInput{
			InternetGatewayId: netStk.Gateway.InternetGatewayId,
			VpcId:             netStk.Vpc.VpcId,
		})
		if err != nil {
			return errors.Wrapf(err, "error detaching internet gateway %s from vpc %s in region %s", *netStk.Gateway.InternetGatewayId, *netStk.Vpc.VpcId, region)
		}

		_, err = sess.DeleteInternetGateway(&ec2.DeleteInternetGatewayInput{
			InternetGatewayId: netStk.Gateway.InternetGatewayId,
		})
		if err != nil {
			return errors.Wrapf(err, "error deleting internget gateway %s in vpc %s in region %s", *netStk.Gateway.InternetGatewayId, *netStk.Vpc.VpcId, region)
		}
	}

	_, err = sess.DeleteVpc(&ec2.DeleteVpcInput{
		VpcId: netStk.Vpc.VpcId,
	})
	if err != nil {
		return errors.Wrapf(err, "error deleting vpc %s in region %s", *netStk.Vpc.VpcId, region)
	}

	return nil
}

func CreateRegionWkspNetworkStack(region string) (*AwsWkspRegionNetworkStack, error) {
	vpcName := fmt.Sprintf(vpcNameFmt, region)
	gatewayName := fmt.Sprintf(gatewayNameFmt, region)
	routeTableName := fmt.Sprintf(routeTableNameFmt, region)
	routeName := fmt.Sprintf(routeNameFmt, region)

	rv := &AwsWkspRegionNetworkStack{}

	sess, err := CreateAwsEc2Session(region)
	if err != nil {
		return rv, errors.Wrapf(err, "error creating ec2 session for region %s", region)
	}

	azsInRegion, err := sess.DescribeAvailabilityZones(&ec2.DescribeAvailabilityZonesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("region-name"),
				Values: []*string{aws.String(region)},
			},
		},
	})
	if err != nil {
		return rv, errors.Wrapf(err, "error getting azs for region %s", region)
	}

	vpc, err := CreateVPC(sess, vpcName, vpcCidr)
	if err != nil {
		return rv, errors.Wrapf(err, "error creating vpc for region %s", region)
	}
	rv.Vpc = vpc

	gateway, err := CreateInternetGateway(sess, gatewayName)
	if err != nil {
		return rv, errors.Wrapf(err, "error creating gateway for region %s", region)
	}
	rv.Gateway = gateway

	err = AttachIntGatewayVpc(sess, vpc, gateway)
	if err != nil {
		return rv, errors.Wrapf(err, "error attaching vpc and internet gateway for region %s vpc %s gateway %s", region, *vpc.VpcId, *gateway.InternetGatewayId)
	}

	routeTable, err := CreateRouteTable(sess, vpc, routeTableName)
	if err != nil {
		return rv, errors.Wrapf(err, "error creating route table for region %s vpc %s", region, *vpc.VpcId)
	}
	rv.RouteTables = []*ec2.RouteTable{routeTable}

	route, err := CreateRoute(sess, routeTable, gateway, routeName)
	if err != nil || !(*route) {
		return rv, errors.Wrapf(err, "error creating route for region %s route table %s gateway %s", region, *routeTable.RouteTableId, *gateway.InternetGatewayId)
	}

	var subnetCidrArr []string
	if len(azsInRegion.AvailabilityZones) <= 4 {
		subnetCidrArr = subnetUpto4Cidr[:]
	} else {
		subnetCidrArr = subnetUpto8Cidr[:]
	}

	rv.Subnets = []*ec2.Subnet{}
	for ind, avblZone := range azsInRegion.AvailabilityZones {
		subnetName := fmt.Sprintf(subnetNameFmt, region, strconv.Itoa(ind))
		subnetAz := avblZone.ZoneName
		subnet, err := CreateSubnetStack(sess, vpc, vpcName, subnetName, subnetCidrArr[ind], *subnetAz, routeTable)
		if err != nil {
			return rv, errors.Wrapf(err, "error creating subnet %s for region %s vpc %s az %s", subnetName, region, *vpc.VpcId, *subnetAz)
		}
		rv.Subnets = append(rv.Subnets, subnet)
	}

	return rv, nil
}

func CreateVPC(sess *ec2.EC2, name string, vpcCidr string) (*ec2.Vpc, error) {
	vpcOut, err := sess.CreateVpc(&ec2.CreateVpcInput{
		CidrBlock: aws.String(vpcCidr),
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String(ec2.ResourceTypeVpc),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String(constants.NAME_LABEL),
						Value: aws.String(name),
					},
					{
						Key:   aws.String(constants.CREATOR_LABEL),
						Value: aws.String(constants.SPAWNER_SERVICE_LABEL),
					},
					{
						Key:   aws.String(constants.PROVISIONER_LABEL),
						Value: aws.String(constants.RANCHER_LABEL),
					},
					{
						Key:   aws.String(constants.NB_TYPE_TAG_KEY),
						Value: aws.String(constants.NB_REGION_WKSP_NETWORK_STK),
					},
				},
			},
		},
	})

	if err != nil {
		return nil, errors.Wrap(err, "error creating aws vpc")
	}

	waitErr := sess.WaitUntilVpcAvailable(&ec2.DescribeVpcsInput{
		VpcIds: []*string{vpcOut.Vpc.VpcId},
	})

	return vpcOut.Vpc, waitErr
}

func CreateInternetGateway(sess *ec2.EC2, name string) (*ec2.InternetGateway, error) {
	intGateOut, err := sess.CreateInternetGateway(&ec2.CreateInternetGatewayInput{
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String(ec2.ResourceTypeInternetGateway),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String(constants.NAME_LABEL),
						Value: aws.String(name),
					},
					{
						Key:   aws.String(constants.CREATOR_LABEL),
						Value: aws.String(constants.SPAWNER_SERVICE_LABEL),
					},
					{
						Key:   aws.String(constants.PROVISIONER_LABEL),
						Value: aws.String(constants.RANCHER_LABEL),
					},
				},
			},
		},
	})

	if err != nil {
		return nil, errors.Wrap(err, "error creating internet gateway")
	}

	return intGateOut.InternetGateway, nil
}

func AttachIntGatewayVpc(sess *ec2.EC2, vpc *ec2.Vpc, intGateway *ec2.InternetGateway) error {
	_, err := sess.AttachInternetGateway(&ec2.AttachInternetGatewayInput{
		InternetGatewayId: intGateway.InternetGatewayId,
		VpcId:             vpc.VpcId,
	})

	if err != nil {
		return errors.Wrap(err, "error attaching internet gateway to VPC")
	}

	return nil
}

func CreateRouteTable(sess *ec2.EC2, vpc *ec2.Vpc, name string) (*ec2.RouteTable, error) {
	routeTableOut, err := sess.CreateRouteTable(&ec2.CreateRouteTableInput{
		VpcId: vpc.VpcId,
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String(ec2.ResourceTypeRouteTable),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String(constants.NAME_LABEL),
						Value: aws.String(name),
					},
					{
						Key:   aws.String(constants.CREATOR_LABEL),
						Value: aws.String(constants.SPAWNER_SERVICE_LABEL),
					},
					{
						Key:   aws.String(constants.PROVISIONER_LABEL),
						Value: aws.String(constants.RANCHER_LABEL),
					},
				},
			},
		},
	})

	if err != nil {
		return nil, errors.Wrap(err, "error creating aws route table")
	}

	return routeTableOut.RouteTable, nil
}

func CreateRoute(sess *ec2.EC2, routeTable *ec2.RouteTable, intGateway *ec2.InternetGateway, name string) (*bool, error) {
	routeOut, err := sess.CreateRoute(&ec2.CreateRouteInput{
		RouteTableId:         routeTable.RouteTableId,
		DestinationCidrBlock: aws.String("0.0.0.0/0"),
		GatewayId:            intGateway.InternetGatewayId,
	})

	if err != nil {
		return nil, errors.Wrap(err, "error creating aws route")
	}

	return routeOut.Return, nil
}

func CreateSubnet(sess *ec2.EC2, vpc *ec2.Vpc, vpcName string, name string, cidrBlock string, avblZone string) (*ec2.Subnet, error) {
	subnetOut, err := sess.CreateSubnet(&ec2.CreateSubnetInput{
		AvailabilityZone: &avblZone,
		CidrBlock:        &cidrBlock,
		VpcId:            vpc.VpcId,
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String(ec2.ResourceTypeSubnet),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String(constants.NAME_LABEL),
						Value: aws.String(name),
					},
					{
						Key:   aws.String(constants.VPC_TAG_KEY),
						Value: aws.String(vpcName),
					},
					{
						Key:   aws.String(constants.CREATOR_LABEL),
						Value: aws.String(constants.SPAWNER_SERVICE_LABEL),
					},
					{
						Key:   aws.String(constants.PROVISIONER_LABEL),
						Value: aws.String(constants.RANCHER_LABEL),
					},
				},
			},
		},
	})

	if err != nil {
		return nil, errors.Wrap(err, "error creating aws subnet")
	}

	waitErr := sess.WaitUntilSubnetAvailable(&ec2.DescribeSubnetsInput{
		SubnetIds: []*string{subnetOut.Subnet.SubnetId},
	})

	return subnetOut.Subnet, waitErr
}

func ModifySubnetMapPublicIp(sess *ec2.EC2, subnet *ec2.Subnet) error {
	_, err := sess.ModifySubnetAttribute(&ec2.ModifySubnetAttributeInput{
		SubnetId:            subnet.SubnetId,
		MapPublicIpOnLaunch: &ec2.AttributeBooleanValue{Value: aws.Bool(true)},
	})

	if err != nil {
		return errors.Wrap(err, "error modifying subnet %s to map public Ipv4")
	}

	return nil
}

func CreateSubnetRouteTblAssn(sess *ec2.EC2, routeTable *ec2.RouteTable, subnet *ec2.Subnet) (*string, error) {
	assnOut, err := sess.AssociateRouteTable(&ec2.AssociateRouteTableInput{
		RouteTableId: routeTable.RouteTableId,
		SubnetId:     subnet.SubnetId,
	})

	if err != nil {
		return nil, errors.Wrap(err, "error creating aws subnet route table association")
	}

	return assnOut.AssociationId, nil
}

func CreateSubnetStack(sess *ec2.EC2, vpc *ec2.Vpc, vpcName string, name string, cidrBlock string, avblZone string, routeTable *ec2.RouteTable) (*ec2.Subnet, error) {
	subnet, err := CreateSubnet(sess, vpc, vpcName, name, cidrBlock, avblZone)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating subnet %s for vpc %s az %s", name, *vpc.VpcId, avblZone)
	}
	err = ModifySubnetMapPublicIp(sess, subnet)
	if err != nil {
		return nil, errors.Wrapf(err, "error modifying subnet %s for vpc %s az %s", *subnet.SubnetId, *vpc.VpcId, avblZone)
	}

	_, err = CreateSubnetRouteTblAssn(sess, routeTable, subnet)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating subnet route table association for subnet %s route table %s", *subnet.SubnetId, *routeTable.RouteTableId)
	}

	return subnet, nil
}
