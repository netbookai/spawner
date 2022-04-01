package aws

import (
	"context"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/common"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/constants"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookdevs/spawnerservice"
)

func (svc AWSController) GetWorkspacesCost(ctx context.Context, req *proto.GetWorkspacesCostRequest) (*proto.GetWorkspacesCostResponse, error) {

	session, err := NewSession(ctx, "", req.GetAccountName())

	if err != nil {
		svc.logger.Errorw("can't start AWS session", "error", err)
		return nil, err
	}

	stsClient := session.getSTSClient()

	callerIdentity, err := stsClient.GetCallerIdentity(nil)

	if err != nil {
		svc.logger.Errorw("failed to get identity", "error", err)
		return nil, err
	}

	accound_id := callerIdentity.Account

	svc.logger.Debugw("fetched accountId", "id", accound_id)

	filter := &costexplorer.Expression{
		And: []*costexplorer.Expression{

			{
				Dimensions: &costexplorer.DimensionValues{
					Key:    aws.String("LINKED_ACCOUNT"),
					Values: aws.StringSlice([]string{*accound_id}),
				},
			},
			{
				Tags: &costexplorer.TagValues{
					Key:    aws.String(constants.WorkspaceId),
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
		svc.logger.Errorw("failed to get cost ", "error", err)
		return nil, err
	}

	costMap := make(map[string]float64)

	var totalCost float64

	for _, resultByTime := range result.ResultsByTime {

		for _, group := range resultByTime.Groups {
			groupKey := ""

			for _, key := range group.Keys {

				if key != nil {
					groupKey += *key
				}

			}

			groupMetric, ok := group.Metrics[req.GetCostType()]
			if !ok {
				groupMetric = &costexplorer.MetricValue{
					Amount: aws.String("0"),
					Unit:   aws.String("USD"),
				}
			}

			floatCost, err := strconv.ParseFloat(*groupMetric.Amount, 64)
			if err != nil {
				svc.logger.Errorw("error converting amount from str to float", "error", err)
				return nil, err
			}
			floatCost = common.RoundTo(floatCost, 4)
			costMap[groupKey] += floatCost
			totalCost += floatCost

		}

	}

	svc.logger.Infow("service-wise cost calculated", "costMap", costMap, "totalCost", totalCost)

	costResponse := &proto.GetWorkspacesCostResponse{
		TotalCost:   totalCost,
		GroupedCost: costMap,
	}

	return costResponse, nil
}
