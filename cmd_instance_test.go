package main

import (
	"encoding/json"
	"testing"

	gomock "github.com/golang/mock/gomock"
	metalcloud "github.com/metalsoft-io/metal-cloud-sdk-go/v2"
	mock_metalcloud "github.com/metalsoft-io/metalcloud-cli/helpers"
	. "github.com/onsi/gomega"
)

func TestInstanceGetCmd(t *testing.T) {
	RegisterTestingT(t)

	ctrl := gomock.NewController(t)
	client := mock_metalcloud.NewMockMetalCloudClient(ctrl)

	instance := metalcloud.Instance{
		InstanceID:      110,
		InstanceArrayID: 10,
		InstanceCredentials: metalcloud.InstanceCredentials{
			SSH: &metalcloud.SSH{
				Username:        "testu",
				InitialPassword: "testp",
				Port:            22,
			},
		},
	}

	iao := metalcloud.InstanceArrayOperation{
		InstanceArraySubdomain:    "tst",
		InstanceArrayID:           10,
		InstanceArrayDeployStatus: "not_started",
		InstanceArrayDeployType:   "edit",
	}

	ia := metalcloud.InstanceArray{
		InstanceArraySubdomain:     "tst",
		InstanceArrayID:            10,
		InstanceArrayOperation:     &iao,
		InstanceArrayServiceStatus: "ordered",
	}
	infra := metalcloud.Infrastructure{
		InfrastructureID:    10,
		InfrastructureLabel: "tsassd",
	}

	client.EXPECT().
		InstanceGet(gomock.Any()).
		Return(&instance, nil).
		AnyTimes()

	client.EXPECT().
		InstanceArrayGet(gomock.Any()).
		Return(&ia, nil).
		AnyTimes()

	client.EXPECT().
		InfrastructureGet(gomock.Any()).
		Return(&infra, nil).
		AnyTimes()

	cmd := MakeCommand(map[string]interface{}{"instance_id": 110, "show_credentials": true})

	ret, err := instanceGetCmd(&cmd, client)
	Expect(err).To(BeNil())
	Expect(ret).To(ContainSubstring("ID"))

}

func TestInstanceServerReplaceCmd(t *testing.T) {
	RegisterTestingT(t)

	ctrl := gomock.NewController(t)
	client := mock_metalcloud.NewMockMetalCloudClient(ctrl)

	instance := metalcloud.Instance{
		InstanceID:      110,
		InstanceArrayID: 10,
		InstanceCredentials: metalcloud.InstanceCredentials{
			SSH: &metalcloud.SSH{
				Username:        "testu",
				InitialPassword: "testp",
				Port:            22,
			},
		},
	}

	ia := metalcloud.InstanceArray{
		InstanceArraySubdomain: "tst",
		InstanceArrayID:        10,
	}

	infra := metalcloud.Infrastructure{
		InfrastructureID:    10,
		InfrastructureLabel: "tsassd",
	}

	client.EXPECT().
		InstanceGet(gomock.Any()).
		Return(&instance, nil).
		AnyTimes()

	client.EXPECT().
		InstanceArrayGet(gomock.Any()).
		Return(&ia, nil).
		AnyTimes()

	client.EXPECT().
		InfrastructureGet(gomock.Any()).
		Return(&infra, nil).
		AnyTimes()

	var server metalcloud.Server
	json.Unmarshal([]byte(_serverFixture2), &server)

	client.EXPECT().
		ServerGet(gomock.Any(), false).
		Return(&server, nil).
		AnyTimes()

	client.EXPECT().
		InstanceServerReplace(110, 100).
		Return(500, nil).
		MinTimes(2)

	cmd := MakeCommand(map[string]interface{}{
		"instance_id": 110,
		"server_id":   100,
		"autoconfirm": true})

	_, err := instanceServerReplaceCmd(&cmd, client)
	Expect(err).To(BeNil())

	cmd = MakeCommand(map[string]interface{}{
		"instance_id":   110,
		"server_id":     100,
		"return_afc_id": true,
		"autoconfirm":   true})

	ret, err := instanceServerReplaceCmd(&cmd, client)
	Expect(err).To(BeNil())
	Expect(ret).To(Equal("500"))

}
