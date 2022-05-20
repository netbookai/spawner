package aws

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pkg/errors"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/labels"
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
	// subnetUpto4Cidr = [4]string{"192.168.0.0/18", "192.168.64.0/18", "192.168.128.0/18", "192.168.192.0/18"}
	// subnetUpto8Cidr = [8]string{"192.168.0.0/19", "192.168.32.0/19", "192.168.64.0/19", "192.168.96.0/19", "192.168.128.0/19", "192.168.160.0/19", "192.168.192.0/19", "192.168.224.0/19"}
	// Using same subnets because running into isses with instances not being available in 4th AZ like us-west-2d
	subnetUpto4Cidr = [4]string{"192.168.0.0/18", "192.168.64.0/18", "192.168.128.0/18"}
	subnetUpto8Cidr = [8]string{"192.168.0.0/18", "192.168.64.0/18", "192.168.128.0/18"}
)

func tagName(k string) *string {
	if k == labels.NameLabel {
		return aws.String(fmt.Sprintf("tag:%s", k))
	}
	return aws.String(fmt.Sprintf("tag:%s", labels.TagKey(k)))
}

func tagValue(val string) []*string {
	return aws.StringSlice([]string{val})
}

func GetRegionWkspNetworkStack(session *Session) (*AwsWkspRegionNetworkStack, error) {
	sess := session.getEC2Client()
	region := session.Region
	vpcName := fmt.Sprintf(vpcNameFmt, region)

	rv := &AwsWkspRegionNetworkStack{}

	vpcOut, err := sess.DescribeVpcs(&ec2.DescribeVpcsInput{
		Filters: []*ec2.Filter{
			{
				Name:   tagName(labels.NameLabel),
				Values: tagValue(vpcName),
			},
			{
				Name:   tagName(labels.CreatorLabel),
				Values: tagValue(labels.SpawnerServiceLabel),
			},
			{
				Name:   tagName(labels.ProvisionerLabel),
				Values: tagValue(labels.SpawnerServiceLabel),
			},
			{
				Name:   tagName(labels.NBTypeTagkey),
				Values: tagValue(labels.NBRegionWkspNetworkStack),
			},
			{
				Name:   tagName(labels.Scope),
				Values: tagValue( /*aws.*/ labels.ScopeTag()),
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
				Values: tagValue(*vpc.VpcId),
			},
			{
				Name:   tagName(labels.Scope),
				Values: tagValue( /*aws.*/ labels.ScopeTag()),
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
				Values: tagValue(*vpc.VpcId),
			},
			{
				Name:   tagName(labels.Scope),
				Values: tagValue( /*aws.*/ labels.ScopeTag()),
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
				Name:   tagName(labels.VpcTagKey),
				Values: tagValue(vpcName),
			},
			{
				Name:   tagName(labels.Scope),
				Values: tagValue( /*aws.*/ labels.ScopeTag()),
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

func DeleteRegionWkspNetworkStack(session *Session, netStk AwsWkspRegionNetworkStack) error {

	client := session.getEC2Client()
	region := session.Region
	var err error
	for _, subn := range netStk.Subnets {
		_, err := client.DeleteSubnet(&ec2.DeleteSubnetInput{
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
			_, err = client.DeleteRouteTable(&ec2.DeleteRouteTableInput{
				RouteTableId: routeTbl.RouteTableId,
			})
			if err != nil {
				return errors.Wrapf(err, "error deleting route table %s in vpc %s in region %s", *routeTbl.RouteTableId, *netStk.Vpc.VpcId, region)
			}
		}
	}

	if netStk.Gateway != nil {
		_, err = client.DetachInternetGateway(&ec2.DetachInternetGatewayInput{
			InternetGatewayId: netStk.Gateway.InternetGatewayId,
			VpcId:             netStk.Vpc.VpcId,
		})
		if err != nil {
			return errors.Wrapf(err, "error detaching internet gateway %s from vpc %s in region %s", *netStk.Gateway.InternetGatewayId, *netStk.Vpc.VpcId, region)
		}

		_, err = client.DeleteInternetGateway(&ec2.DeleteInternetGatewayInput{
			InternetGatewayId: netStk.Gateway.InternetGatewayId,
		})
		if err != nil {
			return errors.Wrapf(err, "error deleting internget gateway %s in vpc %s in region %s", *netStk.Gateway.InternetGatewayId, *netStk.Vpc.VpcId, region)
		}
	}

	if netStk.Vpc != nil {
		_, err = client.DeleteVpc(&ec2.DeleteVpcInput{
			VpcId: netStk.Vpc.VpcId,
		})
		if err != nil {
			return errors.Wrapf(err, "error deleting vpc %s in region %s", *netStk.Vpc.VpcId, region)
		}
	}

	return nil
}

func CreateRegionWkspNetworkStack(session *Session) (*AwsWkspRegionNetworkStack, error) {
	region := session.Region
	client := session.getEC2Client()

	vpcName := fmt.Sprintf(vpcNameFmt, region)
	gatewayName := fmt.Sprintf(gatewayNameFmt, region)
	routeTableName := fmt.Sprintf(routeTableNameFmt, region)
	routeName := fmt.Sprintf(routeNameFmt, region)

	rv := &AwsWkspRegionNetworkStack{}

	azsInRegion, err := client.DescribeAvailabilityZones(&ec2.DescribeAvailabilityZonesInput{
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

	vpc, err := CreateVPC(client, vpcName, vpcCidr)
	if err != nil {
		return rv, errors.Wrapf(err, "error creating vpc for region %s", region)
	}
	rv.Vpc = vpc

	gateway, err := CreateInternetGateway(client, gatewayName)
	if err != nil {
		return rv, errors.Wrapf(err, "error creating gateway for region %s", region)
	}
	rv.Gateway = gateway

	err = AttachIntGatewayVpc(client, vpc, gateway)
	if err != nil {
		return rv, errors.Wrapf(err, "error attaching vpc and internet gateway for region %s vpc %s gateway %s", region, *vpc.VpcId, *gateway.InternetGatewayId)
	}

	routeTable, err := CreateRouteTable(client, vpc, routeTableName)
	if err != nil {
		return rv, errors.Wrapf(err, "error creating route table for region %s vpc %s", region, *vpc.VpcId)
	}
	rv.RouteTables = []*ec2.RouteTable{routeTable}

	route, err := CreateRoute(client, routeTable, gateway, routeName)
	if err != nil || !(*route) {
		return rv, errors.Wrapf(err, "error creating route for region %s route table %s gateway %s", region, *routeTable.RouteTableId, *gateway.InternetGatewayId)
	}

	var subnetCidrArr []string
	if len(azsInRegion.AvailabilityZones) <= 4 {
		subnetCidrArr = subnetUpto4Cidr[:]
	} else {
		subnetCidrArr = subnetUpto8Cidr[:]
	}

	azs := make([]string, 0)
	for _, az := range azsInRegion.AvailabilityZones {
		azs = append(azs, *az.ZoneName)
	}
	// Sort AZ by names
	sort.Strings(azs)

	rv.Subnets = []*ec2.Subnet{}
	for ind, avblZone := range azs {
		// Only considering first 3 AZs per region
		if ind > 2 {
			break
		}
		subnetName := fmt.Sprintf(subnetNameFmt, region, strconv.Itoa(ind))
		subnetAz := avblZone
		subnet, err := CreateSubnetStack(client, vpc, vpcName, subnetName, subnetCidrArr[ind], subnetAz, routeTable)
		if err != nil {
			return rv, errors.Wrapf(err, "error creating subnet %s for region %s vpc %s az %s", subnetName, region, *vpc.VpcId, subnetAz)
		}
		rv.Subnets = append(rv.Subnets, subnet)
	}

	return rv, nil
}

func CreateVPC(client *ec2.EC2, name string, vpcCidr string) (*ec2.Vpc, error) {
	vpcOut, err := client.CreateVpc(&ec2.CreateVpcInput{
		CidrBlock: aws.String(vpcCidr),
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String(ec2.ResourceTypeVpc),
				Tags: []*ec2.Tag{
					{
						Key:   key(labels.NameLabel),
						Value: aws.String(name),
					},
					{
						Key:   key(labels.CreatorLabel),
						Value: aws.String(labels.SpawnerServiceLabel),
					},
					{
						Key:   key(labels.ProvisionerLabel),
						Value: aws.String(labels.SpawnerServiceLabel),
					},
					{
						Key:   key(labels.NBTypeTagkey),
						Value: aws.String(labels.NBRegionWkspNetworkStack),
					},
					{
						Key:   key(labels.Scope),
						Value: aws.String( /*(internal)aws.*/ labels.ScopeTag()),
					},
				},
			},
		},
	})

	if err != nil {
		return nil, errors.Wrap(err, "error creating aws vpc")
	}

	waitErr := client.WaitUntilVpcAvailable(&ec2.DescribeVpcsInput{
		VpcIds: []*string{vpcOut.Vpc.VpcId},
	})

	return vpcOut.Vpc, waitErr
}

func CreateInternetGateway(client *ec2.EC2, name string) (*ec2.InternetGateway, error) {
	intGateOut, err := client.CreateInternetGateway(&ec2.CreateInternetGatewayInput{
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String(ec2.ResourceTypeInternetGateway),
				Tags: []*ec2.Tag{
					{
						Key:   key(labels.NameLabel),
						Value: aws.String(name),
					},
					{
						Key:   key(labels.CreatorLabel),
						Value: aws.String(labels.SpawnerServiceLabel),
					},
					{
						Key:   key(labels.ProvisionerLabel),
						Value: aws.String(labels.SpawnerServiceLabel),
					},
					{
						Key:   key(labels.Scope),
						Value: aws.String( /*(internal)aws.*/ labels.ScopeTag()),
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

func AttachIntGatewayVpc(client *ec2.EC2, vpc *ec2.Vpc, intGateway *ec2.InternetGateway) error {
	_, err := client.AttachInternetGateway(&ec2.AttachInternetGatewayInput{
		InternetGatewayId: intGateway.InternetGatewayId,
		VpcId:             vpc.VpcId,
	})

	if err != nil {
		return errors.Wrap(err, "error attaching internet gateway to VPC")
	}

	return nil
}

func CreateRouteTable(client *ec2.EC2, vpc *ec2.Vpc, name string) (*ec2.RouteTable, error) {
	routeTableOut, err := client.CreateRouteTable(&ec2.CreateRouteTableInput{
		VpcId: vpc.VpcId,
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String(ec2.ResourceTypeRouteTable),
				Tags: []*ec2.Tag{
					{
						Key:   key(labels.NameLabel),
						Value: aws.String(name),
					},
					{
						Key:   key(labels.CreatorLabel),
						Value: aws.String(labels.SpawnerServiceLabel),
					},
					{
						Key:   key(labels.ProvisionerLabel),
						Value: aws.String(labels.SpawnerServiceLabel),
					},
					{
						Key:   key(labels.Scope),
						Value: aws.String( /*(internal)aws.*/ labels.ScopeTag()),
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

func CreateRoute(client *ec2.EC2, routeTable *ec2.RouteTable, intGateway *ec2.InternetGateway, name string) (*bool, error) {
	routeOut, err := client.CreateRoute(&ec2.CreateRouteInput{
		RouteTableId:         routeTable.RouteTableId,
		DestinationCidrBlock: aws.String("0.0.0.0/0"),
		GatewayId:            intGateway.InternetGatewayId,
	})

	if err != nil {
		return nil, errors.Wrap(err, "error creating aws route")
	}

	return routeOut.Return, nil
}

func CreateSubnet(client *ec2.EC2, vpc *ec2.Vpc, vpcName string, name string, cidrBlock string, avblZone string) (*ec2.Subnet, error) {

	subnetOut, err := client.CreateSubnet(&ec2.CreateSubnetInput{
		AvailabilityZone: &avblZone,
		CidrBlock:        &cidrBlock,
		VpcId:            vpc.VpcId,
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String(ec2.ResourceTypeSubnet),
				Tags: []*ec2.Tag{
					{
						Key:   key(labels.NameLabel),
						Value: aws.String(name),
					},
					{
						Key:   key(labels.VpcTagKey),
						Value: aws.String(vpcName),
					},
					{
						Key:   key(labels.CreatorLabel),
						Value: aws.String(labels.SpawnerServiceLabel),
					},
					{
						Key:   key(labels.ProvisionerLabel),
						Value: aws.String(labels.SpawnerServiceLabel),
					},
					{
						Key:   key(labels.Scope),
						Value: aws.String( /*(internal)aws.*/ labels.ScopeTag()),
					},
				},
			},
		},
	})

	if err != nil {
		return nil, errors.Wrap(err, "error creating aws subnet")
	}

	waitErr := client.WaitUntilSubnetAvailable(&ec2.DescribeSubnetsInput{
		SubnetIds: []*string{subnetOut.Subnet.SubnetId},
	})

	return subnetOut.Subnet, waitErr
}

func ModifySubnetMapPublicIp(client *ec2.EC2, subnet *ec2.Subnet) error {
	_, err := client.ModifySubnetAttribute(&ec2.ModifySubnetAttributeInput{
		SubnetId:            subnet.SubnetId,
		MapPublicIpOnLaunch: &ec2.AttributeBooleanValue{Value: aws.Bool(true)},
	})

	if err != nil {
		return errors.Wrap(err, "error modifying subnet %s to map public Ipv4")
	}

	return nil
}

func CreateSubnetRouteTblAssn(client *ec2.EC2, routeTable *ec2.RouteTable, subnet *ec2.Subnet) (*string, error) {
	assnOut, err := client.AssociateRouteTable(&ec2.AssociateRouteTableInput{
		RouteTableId: routeTable.RouteTableId,
		SubnetId:     subnet.SubnetId,
	})

	if err != nil {
		return nil, errors.Wrap(err, "error creating aws subnet route table association")
	}

	return assnOut.AssociationId, nil
}

func CreateSubnetStack(client *ec2.EC2, vpc *ec2.Vpc, vpcName string, name string, cidrBlock string, avblZone string, routeTable *ec2.RouteTable) (*ec2.Subnet, error) {
	subnet, err := CreateSubnet(client, vpc, vpcName, name, cidrBlock, avblZone)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating subnet %s for vpc %s az %s", name, *vpc.VpcId, avblZone)
	}
	err = ModifySubnetMapPublicIp(client, subnet)
	if err != nil {
		return nil, errors.Wrapf(err, "error modifying subnet %s for vpc %s az %s", *subnet.SubnetId, *vpc.VpcId, avblZone)
	}

	_, err = CreateSubnetRouteTblAssn(client, routeTable, subnet)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating subnet route table association for subnet %s route table %s", *subnet.SubnetId, *routeTable.RouteTableId)
	}

	return subnet, nil
}

func key(k string) *string {
	if k == labels.NameLabel {
		return aws.String(k)
	}
	return aws.String(labels.TagKey(k))
}
