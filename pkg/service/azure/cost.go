package azure

import (
	"context"
	"fmt"
	"strconv"

	"github.com/pkg/errors"

	// Note: "github.com/Azure/azure-sdk-for-go/services/costmanagement/mgmt/2019-11-01/costmanagement"
	// using 2019-11-1 version for cost management as latest version had issues
	// before updating in future need to check if the newer version works
	"github.com/Azure/azure-sdk-for-go/services/costmanagement/mgmt/2019-11-01/costmanagement"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/Azure/go-autorest/autorest/to"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/common"
	"gitlab.com/netbook-devs/spawner-service/pkg/service/constants"
	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
)

func (a AzureController) getWorkspacesCost(ctx context.Context, req *proto.GetWorkspacesCostRequest) (*proto.GetWorkspacesCostResponse, error) {

	cred, err := getCredentials(ctx, req.AccountName)
	if err != nil {
		return nil, err
	}

	costClient, err := getCostManagementClient(cred)
	if err != nil {
		return nil, err
	}

	if req.GroupBy.Type != "TAG" {
		a.logger.Errorw("invalid groupby requested", "groupby", req.GroupBy)
		return nil, errors.Wrap(err, "invalid groupby, valid groupby type is TAG")
	}

	grouping, err := getGrouping(req.GroupBy)

	if err != nil {
		a.logger.Errorw("invalid groupby requested", "groupby", req.GroupBy)
		return nil, errors.Wrap(err, "invalid groupby")
	}

	startDate, err := date.ParseTime("2006-01-02", req.StartDate)
	if err != nil {
		a.logger.Errorw("failed to parse start date", "err", err, "req", req)
		return nil, errors.Wrapf(err, "failed to parse start date: %s", req.StartDate)
	}
	endDate, err := date.ParseTime("2006-01-02", req.EndDate)
	if err != nil {
		a.logger.Errorw("failed to parse end date", "req", req)
		return nil, errors.Wrapf(err, "invalid end date: %s", req.EndDate)
	}

	groupFilter := ""

	if req.GroupBy.Type == "TAG" {
		groupFilter = req.GroupBy.Key
	} else {
		groupFilter = constants.WorkspaceId
	}

	scope := "subscriptions/" + cred.SubscriptionID

	result, err := costClient.Usage(ctx, scope, costmanagement.QueryDefinition{
		Type:      costmanagement.ExportTypeActualCost,
		Timeframe: costmanagement.TimeframeTypeCustom,
		TimePeriod: &costmanagement.QueryTimePeriod{
			From: &date.Time{startDate},
			To:   &date.Time{endDate},
		},
		Dataset: &costmanagement.QueryDataset{
			Granularity: costmanagement.GranularityType("None"),
			Grouping:    grouping,
			Filter: &costmanagement.QueryFilter{
				Tags: &costmanagement.QueryComparisonExpression{
					Name:     to.StringPtr(groupFilter),
					Operator: to.StringPtr("In"),
					Values:   &req.WorkspaceIds,
				},
			},
			Aggregation: map[string]*costmanagement.QueryAggregation{
				"totalCostUSD": {
					Name:     to.StringPtr("CostUSD"),
					Function: costmanagement.FunctionTypeSum,
				},
			},
		},
	})

	if err != nil {
		return nil, err
	}

	if result.Response.StatusCode != 200 {
		return nil, fmt.Errorf("azure returned %v status code", result.Response.StatusCode)
	}

	if result.Rows == nil || len(*result.Rows) == 0 {
		a.logger.Infow("didn't find cost for the request", "req", req)
		return &proto.GetWorkspacesCostResponse{}, nil
	}

	var totalCost float64

	groupedCost := make(map[string]float64)

	costUSDColumn := -1
	groupColumn := -1

	if req.GroupBy.Type == "TAG" {

		tagValueColumn := -1

		for i, c := range *result.Columns {
			if *c.Name == constants.CostUSD {
				costUSDColumn = i
			} else if *c.Name == constants.TagValue {
				tagValueColumn = i
			}
		}

		if costUSDColumn == -1 {
			a.logger.Errorw("azure result doesn't have column", "column", constants.CostUSD)
			return nil, fmt.Errorf("azure result doesn't have column: %v", constants.CostUSD)
		}

		if tagValueColumn == -1 {
			a.logger.Errorw("azure result doesn't have column", "column", constants.TagValue)
			return nil, fmt.Errorf("azure result doesn't have column: %v", constants.TagValue)
		}

		groupColumn = tagValueColumn

	} else if req.GroupBy.Key == "SERVICE" {

		serviceNameColumn := -1

		for i, c := range *result.Columns {
			if *c.Name == constants.CostUSD {
				costUSDColumn = i
			} else if *c.Name == constants.ServiceName {
				serviceNameColumn = i
			}
		}

		if costUSDColumn == -1 {
			a.logger.Errorw("azure result doesn't have column", "column", constants.CostUSD)
			return nil, fmt.Errorf("azure result doesn't have column: %v", constants.CostUSD)
		}

		if serviceNameColumn == -1 {
			a.logger.Errorw("azure result doesn't have column", "column", constants.ServiceName)
			return nil, fmt.Errorf("azure result doesn't have column: %v", constants.ServiceName)
		}

		groupColumn = serviceNameColumn
	}

	if groupColumn == -1 {
		a.logger.Errorw("grouping only available for tag and service, couldn't initilize grouping column", "groupBy", req.GroupBy)
		return nil, errors.New("GroupBy only possible for tag and service")
	}

	for _, r := range *result.Rows {

		cost, ok := r[costUSDColumn].(float64)

		if !ok {
			a.logger.Errorw("azure is not returning cost in float")
			return nil, errors.New("failed to parse cost")
		}
		cost = common.RoundTo(cost, 4)

		totalCost += cost

		service, ok := r[groupColumn].(string)
		if !ok {
			a.logger.Error("azure is not returning serviceName in string")
			return nil, errors.New("failed to parse cost")
		}

		groupedCost[service] += cost

	}

	totalCost = common.RoundTo(totalCost, 4)

	costResponse := &proto.GetWorkspacesCostResponse{
		TotalCost:   totalCost,
		GroupedCost: groupedCost,
	}

	return costResponse, nil
}

// getGrouping retruns cost grouping based on azure
func getGrouping(groupBy *proto.GroupBy) (*[]costmanagement.QueryGrouping, error) {

	var grouping *[]costmanagement.QueryGrouping

	if groupBy.Type == "TAG" {

		grouping = &[]costmanagement.QueryGrouping{
			{
				Type: "TagKey",
				Name: &groupBy.Key,
			},
		}
	} else if groupBy.Type == "DIMENSION" && groupBy.Key == "SERVICE" {
		grouping = &[]costmanagement.QueryGrouping{
			{
				Type: costmanagement.QueryColumnTypeDimension,
				Name: to.StringPtr(constants.ServiceName),
			},
		}
	} else {
		return nil, errors.New("invalid groupby requested")
	}

	return grouping, nil

}

func (a AzureController) getCostByTime(ctx context.Context, req *proto.GetCostByTimeRequest) (*proto.GetCostByTimeResponse, error) {

	cred, err := getCredentials(ctx, req.AccountName)
	if err != nil {
		return nil, err
	}

	costClient, err := getCostManagementClient(cred)
	if err != nil {
		return nil, err
	}

	grouping, err := getGrouping(req.GroupBy)

	if err != nil {
		a.logger.Errorw("invalid groupby requested", "groupby", req.GroupBy)
		return nil, errors.Wrap(err, "invalid groupby")
	}

	startDate, err := date.ParseTime("2006-01-02", req.StartDate)
	if err != nil {
		a.logger.Errorw("failed to parse start date", "err", err, "req", req)
		return nil, errors.Wrapf(err, "failed to parse start date: %s", req.StartDate)
	}
	endDate, err := date.ParseTime("2006-01-02", req.EndDate)
	if err != nil {
		a.logger.Errorw("failed to parse end date", "req", req)
		return nil, errors.Wrapf(err, "invalid end date: %s", req.EndDate)
	}

	scope := "subscriptions/" + cred.SubscriptionID

	result, err := costClient.Usage(ctx, scope, costmanagement.QueryDefinition{
		Type:      costmanagement.ExportTypeActualCost,
		Timeframe: costmanagement.TimeframeTypeCustom,
		TimePeriod: &costmanagement.QueryTimePeriod{
			From: &date.Time{startDate},
			To:   &date.Time{endDate},
		},
		Dataset: &costmanagement.QueryDataset{
			Granularity: costmanagement.GranularityTypeDaily,
			Grouping:    grouping,
			Filter: &costmanagement.QueryFilter{
				Tags: &costmanagement.QueryComparisonExpression{
					Name:     to.StringPtr(constants.WorkspaceId),
					Operator: to.StringPtr("In"),
					Values:   &req.Ids,
				},
			},
			Aggregation: map[string]*costmanagement.QueryAggregation{
				"totalCostUSD": {
					Name:     to.StringPtr("CostUSD"),
					Function: costmanagement.FunctionTypeSum,
				},
			},
		},
	})

	if err != nil {
		return nil, err
	}

	if result.Response.StatusCode != 200 {
		return nil, fmt.Errorf("azure returned %v status code", result.Response.StatusCode)
	}

	if result.Rows == nil || len(*result.Rows) == 0 {
		a.logger.Infow("didn't find cost for the request", "req", req)
		return &proto.GetCostByTimeResponse{}, nil
	}

	var totalCost float64

	groupedCost := make(map[string]map[string]float64)

	costUSDColumn := -1
	usageDateColumn := -1
	tagValueColumn := -1

	for i, c := range *result.Columns {
		if *c.Name == constants.CostUSD {
			costUSDColumn = i
		} else if *c.Name == constants.TagValue {
			tagValueColumn = i
		} else if *c.Name == constants.UsageDate {
			usageDateColumn = i
		}
	}

	if costUSDColumn == -1 {
		a.logger.Errorw("azure result doesn't have column", "column", constants.CostUSD)
		return nil, fmt.Errorf("azure result doesn't have column: %v", constants.CostUSD)
	}

	if tagValueColumn == -1 {
		a.logger.Errorw("azure result doesn't have column", "column", constants.TagValue)
		return nil, fmt.Errorf("azure result doesn't have column: %v", constants.TagValue)
	}

	if tagValueColumn == -1 {
		a.logger.Errorw("grouping only available for tag and service, couldn't initilize grouping column", "groupBy", req.GroupBy)
		return nil, errors.New("GroupBy only possible for tag and service")
	}

	for _, r := range *result.Rows {

		cost, ok := r[costUSDColumn].(float64)

		if !ok {
			a.logger.Errorw("azure is not returning cost in float")
			return nil, errors.New("failed to parse cost")
		}
		cost = common.RoundTo(cost, 4)

		totalCost += cost

		service, ok := r[tagValueColumn].(string)
		if !ok {
			a.logger.Error("azure is not returning serviceName in string")
			return nil, errors.New("failed to parse cost")
		}

		if groupedCost[service] == nil {
			groupedCost[service] = make(map[string]float64)
		}

		usageDate, ok := r[usageDateColumn].(float64)
		if !ok {
			a.logger.Error("azure is not returning usageDateColumn")
			return nil, errors.New("failed to parse cost")
		}

		usageDateString := strconv.FormatFloat(usageDate, 'f', -1, 64)

		groupedCost[service][usageDateString] += cost

	}

	resMap := make(map[string]*proto.CostMap)

	for k, v := range groupedCost {

		resMap[k] = &proto.CostMap{
			Cost: v,
		}

	}

	costResponse := &proto.GetCostByTimeResponse{
		GroupedCost: resMap,
	}

	return costResponse, nil
}
