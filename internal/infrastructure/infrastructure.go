package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/metalsoft-io/metalcloud-cli/pkg/api"
	"github.com/metalsoft-io/metalcloud-cli/pkg/formatter"
	"github.com/metalsoft-io/metalcloud-cli/pkg/logger"
	"github.com/metalsoft-io/metalcloud-cli/pkg/response_inspector"
	sdk "github.com/metalsoft-io/metalcloud-sdk-go"
)

var infrastructurePrintConfig = formatter.PrintConfig{
	FieldsConfig: map[string]formatter.RecordFieldConfig{
		"Id": {
			Title: "#",
		},
		"Label": {},
		"ServiceStatus": {
			Title:       "Status",
			Transformer: formatter.FormatStatusValue,
		},
		"UserIdOwner": {
			Title: "Owner",
		},
		"SiteId": {
			Title: "Site",
		},
		"CreatedTimestamp": {
			Title:       "Created",
			Transformer: formatter.FormatDateTimeValue,
		},
		"UpdatedTimestamp": {
			Title:       "Updated",
			Transformer: formatter.FormatDateTimeValue,
		},
	},
}

func InfrastructureList(ctx context.Context, showOwnOnly bool, showOrdered bool, showDeleted bool) error {
	logger.Get().Info().Msgf("Listing all infrastructures")

	client := api.GetApiClient(ctx)

	request := client.InfrastructureAPI.GetInfrastructures(ctx)

	if showOwnOnly {
		userId := api.GetUserId(ctx)
		request = request.FilterUserIdOwner([]string{"$eq:" + userId})
	}

	statusFilters := []string{}
	if !showOrdered {
		statusFilters = append(statusFilters, "$not:$eq:ordered")
	}
	if !showDeleted {
		statusFilters = append(statusFilters, "$not:$eq:deleted")
	}

	if len(statusFilters) > 0 {
		request = request.FilterServiceStatus(statusFilters)
	}

	request = request.SortBy([]string{"id:ASC"})

	infrastructureList, httpRes, err := request.Execute()
	if err := response_inspector.InspectResponse(httpRes, err); err != nil {
		return err
	}

	return formatter.PrintResult(infrastructureList, &infrastructurePrintConfig)
}

func InfrastructureGet(ctx context.Context, infrastructureIdOrLabel string) error {
	logger.Get().Info().Msgf("Get infrastructure '%s'", infrastructureIdOrLabel)

	infrastructureInfo, err := GetInfrastructureByIdOrLabel(ctx, infrastructureIdOrLabel)
	if err != nil {
		return err
	}

	return formatter.PrintResult(infrastructureInfo, &infrastructurePrintConfig)
}

func InfrastructureCreate(ctx context.Context, siteId string, infrastructureLabel string) error {
	logger.Get().Info().Msgf("Create infrastructure '%s'", infrastructureLabel)

	siteIdNumber, err := strconv.ParseFloat(siteId, 32)
	if err != nil {
		err := fmt.Errorf("invalid site ID: '%s'", siteId)
		logger.Get().Error().Err(err).Msg("")
		return err
	}

	createInfrastructure := sdk.InfrastructureCreate{
		Label:  infrastructureLabel,
		SiteId: float32(siteIdNumber),
		Meta:   sdk.NewInfrastructureMeta(),
	}

	client := api.GetApiClient(ctx)

	infrastructureInfo, httpRes, err := client.InfrastructureAPI.CreateInfrastructure(ctx).InfrastructureCreate(createInfrastructure).Execute()
	if err := response_inspector.InspectResponse(httpRes, err); err != nil {
		return err
	}

	return formatter.PrintResult(infrastructureInfo, &infrastructurePrintConfig)
}

func InfrastructureUpdate(ctx context.Context, infrastructureIdOrLabel string, label string, customVariables string) error {
	logger.Get().Info().Msgf("Update infrastructure '%s'", infrastructureIdOrLabel)

	infrastructureInfo, err := GetInfrastructureByIdOrLabel(ctx, infrastructureIdOrLabel)
	if err != nil {
		return err
	}

	updateInfrastructure := sdk.UpdateInfrastructure{}

	if label != "" {
		updateInfrastructure.Label = &label
	} else {
		updateInfrastructure.Label = &infrastructureInfo.Label
	}

	if customVariables != "" {
		err = json.Unmarshal([]byte(customVariables), &updateInfrastructure.CustomVariables)
		if err != nil {
			logger.Get().Error().Err(err).Msg("")
			return err
		}
	}

	client := api.GetApiClient(ctx)

	infrastructureInfo, httpRes, err := client.InfrastructureAPI.UpdateInfrastructureConfiguration(ctx, infrastructureInfo.Id).
		UpdateInfrastructure(updateInfrastructure).
		IfMatch(strconv.Itoa(int(*infrastructureInfo.Config.Revision))).
		Execute()
	if err := response_inspector.InspectResponse(httpRes, err); err != nil {
		return err
	}

	return formatter.PrintResult(infrastructureInfo, &infrastructurePrintConfig)
}

func GetInfrastructureByIdOrLabel(ctx context.Context, infrastructureIdOrLabel string) (*sdk.Infrastructure, error) {
	client := api.GetApiClient(ctx)

	infrastructureList, httpRes, err := client.InfrastructureAPI.GetInfrastructures(ctx).Search(infrastructureIdOrLabel).Execute()
	if err := response_inspector.InspectResponse(httpRes, err); err != nil {
		return nil, err
	}

	if len(infrastructureList.Data) == 0 {
		err := fmt.Errorf("infrastructure '%s' not found", infrastructureIdOrLabel)
		logger.Get().Error().Err(err).Msg("")
		return nil, err
	}

	var infrastructureInfo sdk.Infrastructure
	for _, infrastructure := range infrastructureList.Data {
		if infrastructure.Label == infrastructureIdOrLabel {
			infrastructureInfo = infrastructure
			break
		}

		if strconv.Itoa(int(infrastructure.Id)) == infrastructureIdOrLabel {
			infrastructureInfo = infrastructure
			break
		}
	}

	if infrastructureInfo.Id == 0 {
		err := fmt.Errorf("infrastructure '%s' not found", infrastructureIdOrLabel)
		logger.Get().Error().Err(err).Msg("")
		return nil, err
	}

	return &infrastructureInfo, nil
}
