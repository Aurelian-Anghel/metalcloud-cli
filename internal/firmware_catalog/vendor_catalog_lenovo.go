package firmware_catalog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/metalsoft-io/metalcloud-cli/pkg/logger"
	sdk "github.com/metalsoft-io/metalcloud-sdk-go"
)

const (
	lenovoContentServiceUrl = "https://support.lenovo.com/services/ContentService/"

	lenovoSoftwareUpdateComponentXcc  = "XCC"
	lenovoSoftwareUpdateComponentUefi = "UEFI"
	lenovoSoftwareUpdateComponentLxpm = "LXPM"

	lenovoSoftwareUpdateTypeFix        = "Fix"
	lenovoSoftwareUpdateTypeInstallXML = "InstallXML"
	lenovoSoftwareUpdateTypeReadMe     = "ReadMe"
)

type lenovoSoftwareUpdateFile struct {
	Type        string `json:"Type"`
	Description string `json:"Description"`
	URL         string `json:"URL"`
	FileHash    string `json:"FileHash"`
}

type lenovoSoftwareUpdate struct {
	FixID            string                     `json:"FixID"`
	ComponentID      string                     `json:"ComponentID"`
	Files            []lenovoSoftwareUpdateFile `json:"Files"`
	RequisitesFixIDs []string                   `json:"RequisitesFixIDs"`
	Version          string
	UpdateKey        string
}

type lenovoCatalog struct {
	Data []*lenovoSoftwareUpdate `json:"Data"`
}

func (vc *VendorCatalog) processLenovoCatalog(ctx context.Context) error {
	if len(vc.VendorSystemsFilterEx) == 0 {
		return fmt.Errorf("no vendor systems filter provided")
	}

	for serverType, serverSerialNumber := range vc.VendorSystemsFilterEx {
		catalog, err := vc.readLenovoCatalog(serverType, serverSerialNumber)
		if err != nil {
			return err
		}

		for _, softwareUpdate := range catalog.Data {
			if softwareUpdate.ComponentID != lenovoSoftwareUpdateComponentXcc &&
				softwareUpdate.ComponentID != lenovoSoftwareUpdateComponentUefi &&
				softwareUpdate.ComponentID != lenovoSoftwareUpdateComponentLxpm {
				continue
			}

			downloadUrl := ""
			description := ""
			infoUrl := ""
			for _, file := range softwareUpdate.Files {
				if file.Type == lenovoSoftwareUpdateTypeFix {
					downloadUrl = file.URL
					continue
				}
				if file.Type == lenovoSoftwareUpdateTypeInstallXML {
					description = file.Description
					continue
				}
				if file.Type == lenovoSoftwareUpdateTypeReadMe {
					infoUrl = file.URL
					continue
				}
			}

			if downloadUrl == "" {
				logger.Get().Warn().Msgf("no firmware fix was found for software update %s", softwareUpdate.FixID)
				continue
			}

			if description == "" {
				description = softwareUpdate.FixID
			}

			componentVendorConfiguration := map[string]any{
				"requires": softwareUpdate.RequisitesFixIDs,
			}

			supportedDevices := []map[string]interface{}{
				{
					"type": softwareUpdate.UpdateKey,
				},
			}

			supportedSystems := []map[string]interface{}{
				{
					"machineType":  serverType,
					"serialNumber": serverSerialNumber,
				},
			}

			firmwareBinary := sdk.FirmwareBinary{
				ExternalId:             sdk.PtrString(softwareUpdate.FixID),
				Name:                   description,
				VendorInfoUrl:          &infoUrl,
				VendorDownloadUrl:      downloadUrl,
				CacheDownloadUrl:       nil, //	Will be set after the binary is downloaded
				PackageId:              sdk.PtrString(softwareUpdate.FixID),
				PackageVersion:         sdk.PtrString(softwareUpdate.Version),
				RebootRequired:         true,
				UpdateSeverity:         sdk.FIRMWAREBINARYUPDATESEVERITY_UNKNOWN,
				VendorSupportedDevices: supportedDevices,
				VendorSupportedSystems: supportedSystems,
				VendorReleaseTimestamp: nil,
				Vendor:                 componentVendorConfiguration,
			}

			vc.Binaries = append(vc.Binaries, &firmwareBinary)
		}
	}

	return nil
}

// Search the lenovo support site for the server firmware update information. A JSON response is returned and is saved in the local catalog path folder from the raw config file.
func (vc *VendorCatalog) readLenovoCatalog(machineType string, serialNumber string) (*lenovoCatalog, error) {
	if machineType == "" || serialNumber == "" {
		return nil, fmt.Errorf("machine type and serial number must be specified when searching for a lenovo catalog")
	}

	catalog := lenovoCatalog{}

	localCatalogPath := filepath.Join(vc.VendorLocalCatalogPath, fmt.Sprintf("lenovo_%s_%s.json", machineType, serialNumber))

	fileExists := false
	info, err := os.Stat(localCatalogPath)
	if !os.IsNotExist(err) {
		fileExists = !info.IsDir()
	}

	if fileExists {
		logger.Get().Info().Msgf("Reading local Lenovo catalog %s", localCatalogPath)

		content, err := os.ReadFile(localCatalogPath)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(content, &catalog)
		if err != nil {
			return nil, err
		}
	} else {
		logger.Get().Info().Msgf("Download Lenovo catalog for %s", machineType)

		response, err := downloadLenovoFirmwareUpdates(machineType, serialNumber)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal([]byte(response), &catalog)
		if err != nil {
			return nil, err
		}
	}

	return &catalog, nil
}

func downloadLenovoFirmwareUpdates(machineType string, serialNumber string) (string, error) {
	targetInfos := map[string]string{
		"MachineType":  machineType,
		"SerialNumber": serialNumber,
	}

	searchParams := map[string]interface{}{
		"Category":            "",
		"FixIds":              "",
		"IsIncludeData":       "true",
		"IsIncludeMetaData":   "true",
		"IsIncludeRequisites": "true",
		"IsLatest":            "true",
		"QueryType":           "SUP",
		"SelectSupersedes":    "3",
		"SubmitterName":       "",
		"SubmitterVersion":    "",
		"TargetInfos":         []map[string]string{targetInfos},
		"XmlUpdateType":       "",
	}

	jsonParams, err := json.Marshal(searchParams)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, lenovoContentServiceUrl+"SearchDrivers", bytes.NewBuffer(jsonParams))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := http.DefaultClient

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(responseBody), nil
}
