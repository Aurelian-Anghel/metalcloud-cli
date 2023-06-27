package workflows

import (
	"testing"

	gomock "github.com/golang/mock/gomock"
	metalcloud "github.com/metalsoft-io/metal-cloud-sdk-go/v2"
	mock_metalcloud "github.com/metalsoft-io/metalcloud-cli/helpers"
	"github.com/metalsoft-io/metalcloud-cli/internal/command"
	. "github.com/onsi/gomega"
)

func TestWorkflowsGetCmd(t *testing.T) {
	RegisterTestingT(t)

	ctrl := gomock.NewController(t)
	client := mock_metalcloud.NewMockMetalCloudClient(ctrl)

	wf := metalcloud.Workflow{
		WorkflowID:          10,
		WorkflowLabel:       "test",
		WorkflowDescription: "asdsd",
	}
	vtList := map[string]metalcloud.Workflow{
		"test":  wf,
		"test2": wf,
	}

	client.EXPECT().
		Workflows().
		Return(&vtList, nil).
		AnyTimes()

	client.EXPECT().
		WorkflowGet(10).
		Return(&wf, nil).
		AnyTimes()

	stageDef := metalcloud.StageDefinition{
		StageDefinitionID:    10,
		StageDefinitionLabel: "test",
		StageDefinitionTitle: "Test",
	}

	client.EXPECT().
		StageDefinitionGet(30).
		Return(&stageDef, nil).
		AnyTimes()

	stages := []metalcloud.WorkflowStageDefinitionReference{
		{
			WorkflowStageID:       103,
			WorkflowID:            10,
			StageDefinitionID:     30,
			WorkflowStageRunLevel: 1,
		},
	}

	client.EXPECT().
		WorkflowStages(10).
		Return(&stages, nil).
		AnyTimes()

	format := "json"

	cmd := command.Command{
		Arguments: map[string]interface{}{
			"format":               &format,
			"workflow_id_or_label": &wf.WorkflowID,
		},
	}

	ret, err := workflowGetCmd(&cmd, client)

	Expect(err).To(BeNil())
	Expect(ret).ToNot(BeNil())

}
