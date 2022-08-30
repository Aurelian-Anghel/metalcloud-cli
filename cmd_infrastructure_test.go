package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	gomock "github.com/golang/mock/gomock"
	metalcloud "github.com/metalsoft-io/metal-cloud-sdk-go/v2"
	helper "github.com/metalsoft-io/metalcloud-cli/helpers"
	mock_metalcloud "github.com/metalsoft-io/metalcloud-cli/helpers"
	. "github.com/onsi/gomega"
)

func TestInfrastructureRevertCmd(t *testing.T) {
	RegisterTestingT(t)
	ctrl := gomock.NewController(t)

	infra := metalcloud.Infrastructure{
		InfrastructureID:    10002,
		InfrastructureLabel: "testinfra",
	}

	iao := metalcloud.InstanceArrayOperation{
		InstanceArrayID:           11,
		InstanceArrayLabel:        "testia-edited",
		InstanceArrayDeployType:   "edit",
		InstanceArrayDeployStatus: "not_started",
	}

	ia := metalcloud.InstanceArray{
		InstanceArrayID:            11,
		InstanceArrayLabel:         "testia",
		InfrastructureID:           infra.InfrastructureID,
		InstanceArrayOperation:     &iao,
		InstanceArrayServiceStatus: "active",
	}

	client := helper.NewMockMetalCloudClient(ctrl)

	client.EXPECT().
		InfrastructureGet(infra.InfrastructureID).
		Return(&infra, nil).
		AnyTimes()

	iList := map[string]metalcloud.Infrastructure{
		infra.InfrastructureLabel: infra,
	}

	client.EXPECT().
		Infrastructures().
		Return(&iList, nil).
		AnyTimes()

	client.EXPECT().
		InstanceArrayGet(ia.InstanceArrayID).
		Return(&ia, nil).
		AnyTimes()

	autoconfirm := true
	id := fmt.Sprintf("%d", infra.InfrastructureID)
	cmd := Command{
		Arguments: map[string]interface{}{
			"infrastructure_id_or_label": &id,
			"autoconfirm":                &autoconfirm,
		},
	}

	client.EXPECT().
		InfrastructureOperationCancel(gomock.Any()).
		Return(nil).
		Times(1)

	ret, err := infrastructureRevertCmd(&cmd, client)

	Expect(ret).To(Equal(""))
	Expect(err).To(BeNil())

}

func TestInfrastructureDeployCmd(t *testing.T) {
	RegisterTestingT(t)
	ctrl := gomock.NewController(t)

	infra := metalcloud.Infrastructure{
		InfrastructureID:    10002,
		InfrastructureLabel: "testinfra",
	}

	iao := metalcloud.InstanceArrayOperation{
		InstanceArrayID:           11,
		InstanceArrayLabel:        "testia-edited",
		InstanceArrayDeployType:   "edit",
		InstanceArrayDeployStatus: "not_started",
	}

	ia := metalcloud.InstanceArray{
		InstanceArrayID:            11,
		InstanceArrayLabel:         "testia",
		InfrastructureID:           infra.InfrastructureID,
		InstanceArrayOperation:     &iao,
		InstanceArrayServiceStatus: "active",
	}

	client := helper.NewMockMetalCloudClient(ctrl)

	client.EXPECT().
		InfrastructureGet(10002).
		Return(&infra, nil).
		AnyTimes()

	client.EXPECT().
		InstanceArrayGet(ia.InstanceArrayID).
		Return(&ia, nil).
		AnyTimes()
	//bFalse := true
	bTrue := true
	timeout := 256
	id := fmt.Sprintf("%d", infra.InfrastructureID)
	cmd := Command{
		Arguments: map[string]interface{}{
			"infrastructure_id_or_label":    &id,
			"allow_data_loss":               &bTrue,
			"no_attempt_soft_shutdown":      &bTrue,
			"soft_shutdown_timeout_seconds": &timeout,
		},
	}

	expectedShutdownOptions := metalcloud.ShutdownOptions{
		HardShutdownAfterTimeout:   true,
		AttemptSoftShutdown:        false,
		SoftShutdownTimeoutSeconds: timeout,
	}

	client.EXPECT().
		InfrastructureDeploy(gomock.Any(), expectedShutdownOptions, true, false).
		Return(nil).
		Times(1)

	//test first without confirmation
	ret, err := infrastructureDeployCmd(&cmd, client)
	Expect(ret).To(Equal(""))
	Expect(err).NotTo(BeNil()) //should throw error indicating confirmation not given
	Expect(err.Error()).To(Equal("Operation not confirmed. Aborting"))

	cmd.Arguments["autoconfirm"] = &bTrue

	ret, err = infrastructureDeployCmd(&cmd, client)
	Expect(ret).To(Equal(""))
	Expect(err).To(BeNil()) //should be nil

}

func TestInfrastructureDeleteCmd(t *testing.T) {
	RegisterTestingT(t)
	ctrl := gomock.NewController(t)

	infra := metalcloud.Infrastructure{
		InfrastructureID:    10002,
		InfrastructureLabel: "testinfra",
	}

	iao := metalcloud.InstanceArrayOperation{
		InstanceArrayID:           11,
		InstanceArrayLabel:        "testia-edited",
		InstanceArrayDeployType:   "edit",
		InstanceArrayDeployStatus: "not_started",
	}

	ia := metalcloud.InstanceArray{
		InstanceArrayID:            11,
		InstanceArrayLabel:         "testia",
		InfrastructureID:           infra.InfrastructureID,
		InstanceArrayOperation:     &iao,
		InstanceArrayServiceStatus: "active",
	}

	client := helper.NewMockMetalCloudClient(ctrl)

	client.EXPECT().
		InfrastructureGet(gomock.Any()).
		Return(&infra, nil).
		AnyTimes()

	client.EXPECT().
		InstanceArrayGet(ia.InstanceArrayID).
		Return(&ia, nil).
		AnyTimes()

	id := fmt.Sprintf("%d", infra.InfrastructureID)
	cmd := Command{
		Arguments: map[string]interface{}{
			"infrastructure_id_or_label": &id,
		},
	}

	client.EXPECT().
		InfrastructureDelete(gomock.Any()).
		Return(nil).
		Times(1)

	//test first without confirmation
	ret, err := infrastructureDeleteCmd(&cmd, client)
	Expect(ret).To(Equal(""))
	Expect(err).NotTo(BeNil()) //should throw error indicating confirmation not given
	Expect(err.Error()).To(Equal("Operation not confirmed. Aborting"))

	bTrue := true
	cmd.Arguments["autoconfirm"] = &bTrue

	ret, err = infrastructureDeleteCmd(&cmd, client)
	Expect(ret).To(Equal(""))
	Expect(err).To(BeNil()) //should be nil

	//test with no args
	cmd = Command{
		Arguments: map[string]interface{}{},
	}

	ret, err = infrastructureDeleteCmd(&cmd, client)

	Expect(err).NotTo(BeNil()) //should throw error indicating confirmation not given
}

func TestInfrastructureGetCmd(t *testing.T) {
	RegisterTestingT(t)
	ctrl := gomock.NewController(t)

	io := metalcloud.InfrastructureOperation{
		InfrastructureLabel: "testinfra",
	}

	infra := metalcloud.Infrastructure{
		InfrastructureID:        10002,
		InfrastructureLabel:     "testinfra",
		InfrastructureOperation: io,
	}

	io2 := metalcloud.InfrastructureOperation{
		InfrastructureLabel: "testinfra2",
	}

	infra2 := metalcloud.Infrastructure{
		InfrastructureID:        10003,
		InfrastructureLabel:     "testinfra2",
		InfrastructureOperation: io2,
	}

	iao := metalcloud.InstanceArrayOperation{
		InstanceArrayID:           11,
		InstanceArrayLabel:        "testia-edited",
		InstanceArrayDeployType:   "edit",
		InstanceArrayDeployStatus: "not_started",
	}

	ia := metalcloud.InstanceArray{
		InstanceArrayID:            11,
		InstanceArrayLabel:         "testia",
		InfrastructureID:           infra.InfrastructureID,
		InstanceArrayOperation:     &iao,
		InstanceArrayServiceStatus: "active",
	}

	dao := metalcloud.DriveArrayOperation{
		DriveArrayID:           10,
		DriveArrayLabel:        "test-edited",
		InstanceArrayID:        ia.InstanceArrayID,
		InfrastructureID:       infra.InfrastructureID,
		DriveArrayCount:        101,
		DriveArrayDeployType:   "edit",
		DriveArrayDeployStatus: "not_started",
	}

	da := metalcloud.DriveArray{
		DriveArrayID:            10,
		DriveArrayLabel:         "test",
		InstanceArrayID:         ia.InstanceArrayID,
		InfrastructureID:        infra.InfrastructureID,
		DriveArrayCount:         101,
		DriveArrayOperation:     &dao,
		DriveArrayServiceStatus: "active",
	}

	client := helper.NewMockMetalCloudClient(ctrl)

	//given this return the other
	infraListByID := map[interface{}]*metalcloud.Infrastructure{
		infra.InfrastructureID:  &infra,
		infra2.InfrastructureID: &infra,
	}

	client.EXPECT().
		InfrastructureGet(gomock.Any()).
		DoAndReturn(
			func(i int) (*metalcloud.Infrastructure, error) {
				if intf, ok := infraListByID[i]; ok {
					return intf, nil
				}
				return nil, fmt.Errorf("could not find infra with id %v", i)
			}).
		AnyTimes()

	infraListByLabel := map[interface{}]*metalcloud.Infrastructure{
		infra.InfrastructureLabel:  &infra,
		infra2.InfrastructureLabel: &infra2,
	}

	client.EXPECT().
		InfrastructureGetByLabel(gomock.Any()).
		DoAndReturn(
			func(label string) (*metalcloud.Infrastructure, error) {
				if intf, ok := infraListByLabel[label]; ok {
					return intf, nil
				}
				return nil, fmt.Errorf("could not find infra with label %v", label)
			}).
		AnyTimes()

	client.EXPECT().
		InstanceArrayGet(ia.InstanceArrayID).
		Return(&ia, nil).
		AnyTimes()

	format := "json"
	cmd := Command{
		Arguments: map[string]interface{}{
			"infrastructure_id_or_label": &infra.InfrastructureID,
			"format":                     &format,
		},
	}

	iaList := map[string]metalcloud.InstanceArray{
		ia.InstanceArrayLabel + ".vanilla": ia,
	}

	client.EXPECT().
		InstanceArrays(gomock.Any()).
		Return(&iaList, nil).
		AnyTimes()

	daList := map[string]metalcloud.DriveArray{
		da.DriveArrayLabel + ".vanilla": da,
	}
	client.EXPECT().
		DriveArrays(gomock.Any()).
		Return(&daList, nil).
		AnyTimes()

	sda := metalcloud.SharedDrive{
		SharedDriveLabel: "test",
	}
	sdaList := map[string]metalcloud.SharedDrive{
		sda.SharedDriveLabel + ".vanilla": sda,
	}
	client.EXPECT().
		SharedDrives(gomock.Any()).
		Return(&sdaList, nil).
		AnyTimes()

	ret, err := infrastructureGetCmd(&cmd, client)
	Expect(err).To(BeNil())
	Expect(ret).To(Not(Equal("")))

	var m []interface{}
	err = json.Unmarshal([]byte(ret), &m)

	Expect(err).To(BeNil())

	r := m[0].(map[string]interface{})
	Expect(r["LABEL"].(string)).To(Equal(iao.InstanceArrayLabel))

	//test with label instead of id

	cmd = Command{
		Arguments: map[string]interface{}{
			"infrastructure_id_or_label": &infra.InfrastructureLabel,
			"format":                     &format,
		},
	}

	ret, err = infrastructureGetCmd(&cmd, client)
	Expect(err).To(BeNil())

	err = json.Unmarshal([]byte(ret), &m)

	Expect(err).To(BeNil())
}

func TestInfrastructureListAdminCmd(t *testing.T) {
	RegisterTestingT(t)
	ctrl := gomock.NewController(t)

	infra := metalcloud.InfrastructuresSearchResult{
		InfrastructureID:            10002,
		InfrastructureLabel:         "testinfra",
		InfrastructureServiceStatus: "active",
	}

	infra2 := metalcloud.InfrastructuresSearchResult{
		InfrastructureID:            10003,
		InfrastructureLabel:         "testinfra2",
		InfrastructureServiceStatus: "ordered",
	}
	infra3 := metalcloud.InfrastructuresSearchResult{
		InfrastructureID:            10003,
		InfrastructureLabel:         "testinfra3",
		InfrastructureServiceStatus: "active",
		AFCExecutedSuccess:          10,
		AFCTotal:                    10,
	}

	infra4 := metalcloud.InfrastructuresSearchResult{
		InfrastructureID:            10003,
		InfrastructureLabel:         "testinfra4",
		InfrastructureServiceStatus: "active",
		AFCExecutedSuccess:          9,
		AFCThrownError:              1,
		AFCTotal:                    10,
	}

	infraList := []metalcloud.InfrastructuresSearchResult{
		infra,
		infra2,
		infra3,
		infra4,
	}

	client := helper.NewMockMetalCloudClient(ctrl)

	client.EXPECT().
		InfrastructureSearch(gomock.Any()).
		Return(&infraList, nil).
		AnyTimes()

	//test plaintext return
	format := ""
	cmd := Command{
		Arguments: map[string]interface{}{
			"format": &format,
		},
	}

	ret, err := infrastructureListAdminCmd(&cmd, client)
	Expect(err).To(BeNil())
	Expect(ret).To(Not(Equal("")))
	Expect(ret).To(ContainSubstring(infra.InfrastructureLabel))
	Expect(ret).NotTo(ContainSubstring(infra2.InfrastructureLabel)) //does not show ordered infra by default
	Expect(ret).To(ContainSubstring(infra3.InfrastructureLabel))
	Expect(ret).To(ContainSubstring("Deploy ongoing - Thrown error at 9/10"))

	//test json return
	format = "json"
	cmd.Arguments["format"] = &format

	ret, err = infrastructureListAdminCmd(&cmd, client)
	Expect(err).To(BeNil())
	Expect(ret).To(Not(Equal("")))

	var m []interface{}
	err = json.Unmarshal([]byte(ret), &m)
	Expect(err).To(BeNil())

	//test csv return
	format = "csv"
	cmd.Arguments["format"] = &format

	ret, err = infrastructureListAdminCmd(&cmd, client)
	Expect(err).To(BeNil())
	Expect(ret).To(Not(Equal("")))

	reader := csv.NewReader(strings.NewReader(ret))

	csv, err := reader.ReadAll()
	Expect(err).To(BeNil())
	Expect(csv).NotTo(BeNil())

}

func TestInfrastructureListUserCmd(t *testing.T) {
	RegisterTestingT(t)
	ctrl := gomock.NewController(t)

	infra := metalcloud.Infrastructure{
		InfrastructureID:            10002,
		InfrastructureLabel:         "testinfra",
		InfrastructureServiceStatus: "active",
	}

	infra2 := metalcloud.Infrastructure{
		InfrastructureID:            10003,
		InfrastructureLabel:         "testinfra2",
		InfrastructureServiceStatus: "ordered",
	}
	infra3 := metalcloud.Infrastructure{
		InfrastructureID:            10003,
		InfrastructureLabel:         "testinfra3",
		InfrastructureServiceStatus: "active",
	}

	infra4 := metalcloud.Infrastructure{
		InfrastructureID:            10003,
		InfrastructureLabel:         "testinfra4",
		InfrastructureServiceStatus: "active",
	}

	infraList := map[string]metalcloud.Infrastructure{

		"infra1": infra,
		"infra2": infra2,
		"infra3": infra3,
		"infra4": infra4,
	}

	client := helper.NewMockMetalCloudClient(ctrl)

	client.EXPECT().
		Infrastructures().
		Return(&infraList, nil).
		AnyTimes()

	//test plaintext return
	format := ""
	cmd := Command{
		Arguments: map[string]interface{}{
			"format": &format,
		},
	}

	ret, err := infrastructureListUserCmd(&cmd, client)
	Expect(err).To(BeNil())
	Expect(ret).To(Not(Equal("")))
	Expect(ret).To(ContainSubstring(infra.InfrastructureLabel))
	Expect(ret).To(ContainSubstring(infra2.InfrastructureLabel)) //shows ordered infra by default to avoid user confusion
	Expect(ret).To(ContainSubstring(infra3.InfrastructureLabel))

	//test json return
	format = "json"
	cmd.Arguments["format"] = &format

	ret, err = infrastructureListUserCmd(&cmd, client)
	Expect(err).To(BeNil())
	Expect(ret).To(Not(Equal("")))

	var m []interface{}
	err = json.Unmarshal([]byte(ret), &m)
	Expect(err).To(BeNil())

	//test csv return
	format = "csv"
	cmd.Arguments["format"] = &format

	ret, err = infrastructureListUserCmd(&cmd, client)
	Expect(err).To(BeNil())
	Expect(ret).To(Not(Equal("")))

	reader := csv.NewReader(strings.NewReader(ret))

	csv, err := reader.ReadAll()
	Expect(err).To(BeNil())
	Expect(csv).NotTo(BeNil())

}

func TestDeployBlocking(t *testing.T) {
	RegisterTestingT(t)

	ctrl := gomock.NewController(t)
	client := mock_metalcloud.NewMockMetalCloudClient(ctrl)

	instance := metalcloud.Instance{
		InstanceArrayID: 100,
	}

	instanceArray := metalcloud.InstanceArray{
		InfrastructureID: 1000,
	}

	client.EXPECT().
		InstanceGet(10).
		Return(&instance, nil).
		AnyTimes()

	client.EXPECT().
		InstanceArrayGet(100).
		Return(&instanceArray, nil).
		AnyTimes()

	//infrastructureGet first returns locked then returns not locked
	gomock.InOrder(
		client.EXPECT().
			InfrastructureGet(1000).
			Return(&metalcloud.Infrastructure{
				InfrastructureID: 1000,
				InfrastructureOperation: metalcloud.InfrastructureOperation{
					InfrastructureDeployStatus: "ongoing",
				}, //locked infra
			}, nil).
			Times(1),
		client.EXPECT().
			InfrastructureGet(1000).
			Return(&metalcloud.Infrastructure{
				InfrastructureID: 1000,
				InfrastructureOperation: metalcloud.InfrastructureOperation{
					InfrastructureDeployStatus: "finished",
				}, //locked infra
			}, nil).
			Times(1),
	)

	client.EXPECT().
		InfrastructureDeploy(1000, gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()

	cmd := MakeCommand(map[string]interface{}{
		"infrastructure_id_or_label": "1000",
		"autoconfirm":                true,
		"block_until_deployed":       true,
		"block_timeout":              3, //3 seconds to make sure the test is short
		"block_check_interval":       1, //1 second to make sure the test is short
	})

	//cs with infra locked
	_, err := infrastructureDeployCmd(&cmd, client)
	Expect(err).To(BeNil())

}

func TestDeployBlockingTimeouting(t *testing.T) {
	RegisterTestingT(t)

	ctrl := gomock.NewController(t)
	client := mock_metalcloud.NewMockMetalCloudClient(ctrl)

	instance := metalcloud.Instance{
		InstanceArrayID: 100,
	}

	instanceArray := metalcloud.InstanceArray{
		InfrastructureID: 1000,
	}

	client.EXPECT().
		InstanceGet(10).
		Return(&instance, nil).
		AnyTimes()

	client.EXPECT().
		InstanceArrayGet(100).
		Return(&instanceArray, nil).
		AnyTimes()

	client.EXPECT().
		InfrastructureGet(1000).
		Return(&metalcloud.Infrastructure{
			InfrastructureID: 1000,
			InfrastructureOperation: metalcloud.InfrastructureOperation{
				InfrastructureDeployStatus: "ongoing",
			},
		}, nil).
		AnyTimes()

	client.EXPECT().
		InfrastructureDeploy(1000, gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()

	cmd := MakeCommand(map[string]interface{}{
		"infrastructure_id_or_label": "1000",
		"autoconfirm":                true,
		"block_until_deployed":       true,
		"block_timeout":              2, //2 seconds to make sure the test is short
		"block_check_interval":       1, //1 second to make sure the test is short
	})

	//cs with infra locked
	_, err := infrastructureDeployCmd(&cmd, client)
	Expect(err).NotTo(BeNil())

}
