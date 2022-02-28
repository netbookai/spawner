package aws

import (
	"context"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	"gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice/common"
	"gitlab.com/netbook-devs/spawner-service/pkg/spawnerservice/constants"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookdevs/spawnerservice"
)

func (svc AWSController) GetWorkspaceCost(ctx context.Context, req *proto.GetWorkspaceCostRequest) (*proto.GetWorkspaceCostResponse, error) {

	filter := &costexplorer.Expression{
		And: []*costexplorer.Expression{

			{
				Dimensions: &costexplorer.DimensionValues{
					Key:    aws.String("LINKED_ACCOUNT"),
					Values: aws.StringSlice([]string{req.GetAccountName()}),
				},
			},
			{
				Tags: &costexplorer.TagValues{
					Key:    aws.String(constants.WorkspaceId),
					Values: aws.StringSlice([]string{req.GetWorkspaceId()}),
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
				Type: aws.String("DIMENSION"),
				Key:  aws.String(req.GroupBy),
			},
		},
		Metrics: aws.StringSlice([]string{req.GetCostType()}),
		Filter:  filter,
	}

	session, err := NewSession(svc.config, "", req.GetAccountName())

	if err != nil {
		svc.logger.Errorw("can't start AWS session", "error", err)
		return nil, err
	}

	client := session.getCostExplorerClient()

	result, err := client.GetCostAndUsage(&input)

	if err != nil {
		svc.logger.Errorw("failed to get cost ", "error", err)
		return nil, err
	}

	costResponse := &proto.GetWorkspaceCostResponse{}

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

	svc.logger.Infow("service wise cost calculated", "costMap", costMap, "totalCost", totalCost)

	costResponse.TotalCost = totalCost

	for k, v := range costMap {
		costResponse.GroupedCost = append(costResponse.GroupedCost, &proto.GetWorkspaceCostGroupResponse{
			Group: k,
			Cost:  v,
		})
	}

	return costResponse, nil
}
