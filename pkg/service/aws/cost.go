package aws

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	"github.com/pkg/errors"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/common"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/constants"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"

	"github.com/shopspring/decimal"
)

func (svc AWSController) GetWorkspacesCost(ctx context.Context, req *proto.GetWorkspacesCostRequest) (*proto.GetWorkspacesCostResponse, error) {

	session, err := NewSession(ctx, "", req.GetAccountName())

	if err != nil {
		svc.logger.Error(ctx, "can't start AWS session", "error", err)
		return nil, errors.Wrap(err, "GetWorkspacesCost: failed to get aws session")
	}

	stsClient := session.getSTSClient()

	callerIdentity, err := stsClient.GetCallerIdentity(nil)

	if err != nil {
		svc.logger.Error(ctx, "failed to get identity", "error", err)
		return nil, errors.Wrap(err, "GetWorkspacesCost: failed to get callerIdentity")
	}

	account_id := callerIdentity.Account
	svc.logger.Debug(ctx, "fetched accountId", "id", account_id)

	groupFilter := ""

	if req.GroupBy.Type == "TAG" {
		groupFilter = req.GroupBy.Key
	} else {
		groupFilter = constants.WorkspaceId
	}

	filter := &costexplorer.Expression{
		And: []*costexplorer.Expression{

			{
				Dimensions: &costexplorer.DimensionValues{
					Key:    aws.String("LINKED_ACCOUNT"),
					Values: aws.StringSlice([]string{*account_id}),
				},
			},
			{
				Tags: &costexplorer.TagValues{
					Key:    aws.String(groupFilter),
					Values: aws.StringSlice(req.GetWorkspaceIds()),
				},
			},

			{

				Not: &costexplorer.Expression{
					Dimensions: &costexplorer.DimensionValues{
						Key: aws.String("RECORD_TYPE"),
						Values: aws.StringSlice([]string{
							"Credit",
						}),
					},
				},
			},
		},
	}

	input := costexplorer.GetCostAndUsageInput{
		TimePeriod: &costexplorer.DateInterval{
			Start: aws.String(req.GetStartDate()),
			End:   aws.String(req.EndDate),
		},
		Granularity: aws.String(req.GetGranularity()),
		GroupBy: []*costexplorer.GroupDefinition{
			{
				Type: aws.String(req.GroupBy.Type),
				Key:  aws.String(req.GroupBy.Key),
			},
		},
		Metrics: aws.StringSlice([]string{req.GetCostType()}),
		Filter:  filter,
	}

	client := session.getCostExplorerClient()

	result, err := client.GetCostAndUsage(&input)

	if err != nil {
		svc.logger.Error(ctx, "failed to get cost ", "error", err)
		return nil, errors.Wrap(err, "GetWorkspacesCost: failed to get from aws")
	}

	costMap := make(map[string]decimal.Decimal)

	var totalCost decimal.Decimal

	for _, resultByTime := range result.ResultsByTime {

		for _, group := range resultByTime.Groups {
			groupKey := ""

			for _, key := range group.Keys {

				if key != nil {
					key := strings.Split(*key, "$")[1]
					groupKey += key
				}
			}

			groupMetric, ok := group.Metrics[req.GetCostType()]
			if !ok {
				groupMetric = &costexplorer.MetricValue{
					Amount: aws.String("0"),
					Unit:   aws.String("USD"),
				}
			}

			decimalCost, err := decimal.NewFromString(*groupMetric.Amount)
			if err != nil {
				svc.logger.Error(ctx, "error converting amount from str to float", "error", err)
				return nil, errors.Wrap(err, "GetWorkspacesCost: failed to convert amount to decimal")
			}

			costMap[groupKey] = costMap[groupKey].Add(decimalCost)
			totalCost = totalCost.Add(decimalCost)

		}

	}

	svc.logger.Info(ctx, "service-wise cost calculated", "costMap", costMap, "totalCost", totalCost)

	// // skipping bool check if value in decimal and float are exactly same
	// totalCostInFloat, _ := totalCost.Float64()

	totalCostIn100thCents := common.Get100thOfCentsInIntegerForDollar(totalCost)

	costMapInt, err := common.ConverDecimalCostMapToIntCostMap(costMap)
	if err != nil {
		svc.logger.Error(ctx, "failed to convert cost from decimal to int", "error", err)
		return nil, errors.Wrap(err, "GetWorkspacesCost: failed to convert cost to integer")
	}

	costResponse := &proto.GetWorkspacesCostResponse{
		TotalCost:   totalCostIn100thCents,
		GroupedCost: costMapInt,
	}

	return costResponse, nil
}

func getCostAndUsageRequest(account_id *string, req *proto.GetApplicationsCostRequest) costexplorer.GetCostAndUsageInput {

	filter := &costexplorer.Expression{
		And: []*costexplorer.Expression{

			{
				Dimensions: &costexplorer.DimensionValues{
					Key:    aws.String("LINKED_ACCOUNT"),
					Values: aws.StringSlice([]string{*account_id}),
				},
			},
			{
				Tags: &costexplorer.TagValues{
					Key:    aws.String(constants.WorkspaceId),
					Values: aws.StringSlice(req.GetApplicationIds()),
				},
			},

			{

				Not: &costexplorer.Expression{
					Dimensions: &costexplorer.DimensionValues{
						Key: aws.String("RECORD_TYPE"),
						Values: aws.StringSlice([]string{
							"Credit",
						}),
					},
				},
			},
		},
	}

	input := costexplorer.GetCostAndUsageInput{
		TimePeriod: &costexplorer.DateInterval{
			Start: aws.String(req.GetStartDate()),
			End:   aws.String(req.EndDate),
		},
		Granularity: aws.String(req.GetGranularity()),
		GroupBy: []*costexplorer.GroupDefinition{
			{
				Type: aws.String(req.GroupBy.Type),
				Key:  aws.String(req.GroupBy.Key),
			},
		},
		Metrics: aws.StringSlice([]string{req.GetCostType()}),
		Filter:  filter,
	}

	return input
}

func getCostMap(result *costexplorer.GetCostAndUsageOutput, costType string) (map[string]decimal.Decimal, error) {

	costMap := make(map[string]decimal.Decimal)

	for _, resultByTime := range result.ResultsByTime {

		for _, group := range resultByTime.Groups {
			groupKey := ""

			for _, key := range group.Keys {

				if key != nil {
					key := strings.Split(*key, "$")[1]
					groupKey += key
				}
			}

			groupMetric, ok := group.Metrics[costType]
			if !ok {
				groupMetric = &costexplorer.MetricValue{
					Amount: aws.String("0"),
					Unit:   aws.String("USD"),
				}
			}

			decimalCost, err := decimal.NewFromString(*groupMetric.Amount)
			if err != nil {
				return nil, errors.Wrap(err, "GetWorkspacesCost: failed to convert amount to decimal")
			}

			costMap[groupKey] = costMap[groupKey].Add(decimalCost)

		}

	}

	return costMap, nil
}
func (svc AWSController) GetApplicationsCost(ctx context.Context, req *proto.GetApplicationsCostRequest) (*proto.GetApplicationsCostResponse, error) {
	session, err := NewSession(ctx, "", req.GetAccountName())

	if err != nil {
		svc.logger.Error(ctx, "can't start AWS session", "error", err)
		return nil, err
	}

	stsClient := session.getSTSClient()

	callerIdentity, err := stsClient.GetCallerIdentity(nil)

	if err != nil {
		svc.logger.Error(ctx, "failed to get identity", "error", err)
		return nil, err
	}

	account_id := callerIdentity.Account

	svc.logger.Debug(ctx, "fetched accountId", "id", account_id)

	input := getCostAndUsageRequest(account_id, req)

	client := session.getCostExplorerClient()

	result, err := client.GetCostAndUsage(&input)

	if err != nil {
		svc.logger.Error(ctx, "failed to get cost ", "error", err)
		return nil, err
	}

	var totalCost decimal.Decimal

	costMap, err := getCostMap(result, req.GetCostType())

	if err != nil {
		svc.logger.Error(ctx, "failed to get costmap", "error", err)
		return nil, err
	}

	for _, val := range costMap {
		totalCost = totalCost.Add(val)
	}

	svc.logger.Info(ctx, "service-wise cost calculated", "costMap", costMap, "totalCost", totalCost)

	// // skipping bool check if value in decimal and float are exactly same
	// totalCostInFloat, _ := totalCost.Float64()

	totalCostIn100thCents := common.Get100thOfCentsInIntegerForDollar(totalCost)

	costMapInt, err := common.ConverDecimalCostMapToIntCostMap(costMap)
	if err != nil {
		svc.logger.Error(ctx, "failed to convert cost from decimal to int", "error", err)
		return nil, errors.Wrap(err, "GetWorkspacesCost: failed to convert cost to integer")
	}

	costResponse := &proto.GetApplicationsCostResponse{
		TotalCost:   totalCostIn100thCents,
		GroupedCost: costMapInt,
	}

	return costResponse, nil
}

func (svc AWSController) GetCostByTime(ctx context.Context, req *proto.GetCostByTimeRequest) (*proto.GetCostByTimeResponse, error) {

	session, err := NewSession(ctx, "", req.GetAccountName())

	if err != nil {
		svc.logger.Error(ctx, "can't start AWS session", "error", err)
		return nil, errors.Wrap(err, "GetCostByTime: failed to get aws session")
	}

	stsClient := session.getSTSClient()

	callerIdentity, err := stsClient.GetCallerIdentity(nil)

	if err != nil {
		svc.logger.Error(ctx, "failed to get identity", "error", err)
		return nil, errors.Wrap(err, "GetCostByTime: failed to get aws callerIdentity ")
	}

	account_id := callerIdentity.Account

	svc.logger.Debug(ctx, "fetched accountId", "id", account_id)

	filter := &costexplorer.Expression{
		And: []*costexplorer.Expression{

			{
				Dimensions: &costexplorer.DimensionValues{
					Key:    aws.String("LINKED_ACCOUNT"),
					Values: aws.StringSlice([]string{*account_id}),
				},
			},
			{
				Tags: &costexplorer.TagValues{
					Key:    aws.String(req.GroupBy.Key),
					Values: aws.StringSlice(req.Ids),
				},
			},

			{

				Not: &costexplorer.Expression{
					Dimensions: &costexplorer.DimensionValues{
						Key: aws.String("RECORD_TYPE"),
						Values: aws.StringSlice([]string{
							"Credit",
						}),
					},
				},
			},
		},
	}

	input := costexplorer.GetCostAndUsageInput{
		TimePeriod: &costexplorer.DateInterval{
			Start: aws.String(req.StartDate),
			End:   aws.String(req.EndDate),
		},
		Granularity: aws.String(req.GetGranularity()),
		GroupBy: []*costexplorer.GroupDefinition{
			{
				Type: aws.String(req.GroupBy.Type),
				Key:  aws.String(req.GroupBy.Key),
			},
		},
		Metrics: aws.StringSlice([]string{"BlendedCost"}),
		Filter:  filter,
	}

	client := session.getCostExplorerClient()

	result, err := client.GetCostAndUsage(&input)

	if err != nil {
		svc.logger.Error(ctx, "failed to get cost from aws ", "error", err)
		return nil, errors.Wrap(err, "GetCostByTime:  failed to get cost from aws")
	}

	costMap := make(map[string]map[string]decimal.Decimal)

	var totalCost decimal.Decimal

	for _, resultByTime := range result.ResultsByTime {

		for _, group := range resultByTime.Groups {
			groupKey := ""

			for _, key := range group.Keys {

				if key != nil {
					key := strings.Split(*key, "$")[1]
					groupKey += key
				}

			}

			groupMetric, ok := group.Metrics["BlendedCost"]
			if !ok {
				groupMetric = &costexplorer.MetricValue{
					Amount: aws.String("0"),
					Unit:   aws.String("USD"),
				}
			}

			costInDecimal, err := decimal.NewFromString(*groupMetric.Amount)
			if err != nil {
				svc.logger.Error(ctx, "error converting amount from str to decimal", "error", err)
				return nil, err
			}

			if costMap[groupKey] == nil {
				costMap[groupKey] = make(map[string]decimal.Decimal)
			}

			date := strings.ReplaceAll(*resultByTime.TimePeriod.Start, "-", "")

			costMap[groupKey][date] = costMap[groupKey][date].Add(costInDecimal)
			totalCost = totalCost.Add(costInDecimal)

		}

	}

	svc.logger.Info(ctx, "cost calculated", "costMap", costMap)

	costMapInt, err := common.ConverDecimalCostMapOfMapToIntCostMapOfMap(costMap)
	if err != nil {
		svc.logger.Error(ctx, "failed to convert cost from decimal to int", "error", err)
		return nil, errors.Wrap(err, "GetCostByTime: failed to convert cost to integer ")
	}

	resMap := make(map[string]*proto.CostMap)

	for k, v := range costMapInt {

		resMap[k] = &proto.CostMap{
			Cost: v,
		}

	}

	costResponse := &proto.GetCostByTimeResponse{
		GroupedCost: resMap,
	}

	return costResponse, nil
}
