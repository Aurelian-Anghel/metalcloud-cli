package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	metalcloud "github.com/metalsoft-io/metal-cloud-sdk-go/v2"
	"github.com/metalsoft-io/tableformatter"
)

var workflowCmds = []Command{

	{
		Description:  "Lists available workflows.",
		Subject:      "workflow",
		AltSubject:   "wf",
		Predicate:    "list",
		AltPredicate: "ls",
		FlagSet:      flag.NewFlagSet("list workflows", flag.ExitOnError),
		InitFunc: func(c *Command) {
			c.Arguments = map[string]interface{}{
				"usage":  c.FlagSet.String("usage", _nilDefaultStr, "Workflow usage. One of infrastructure, network_equipment, server, free_standing, storage_pool, user, os_template"),
				"format": c.FlagSet.String("format", _nilDefaultStr, "The output format. Supported values are 'json','csv','yaml'. The default format is human readable."),
			}
		},
		ExecuteFunc:   workflowsListCmd,
		Endpoint:      ExtendedEndpoint,
		AdminEndpoint: DeveloperEndpoint,
	},
	{
		Description:  "Get workflow details.",
		Subject:      "workflow",
		AltSubject:   "wf",
		Predicate:    "get",
		AltPredicate: "show",
		FlagSet:      flag.NewFlagSet("list workflows", flag.ExitOnError),
		InitFunc: func(c *Command) {
			c.Arguments = map[string]interface{}{
				"workflow_id_or_label": c.FlagSet.String("id", _nilDefaultStr, "Workflow's id or label."),
				"format":               c.FlagSet.String("format", _nilDefaultStr, "The output format. Supported values are 'json','csv','yaml'. The default format is human readable."),
			}
		},
		ExecuteFunc:   workflowGetCmd,
		Endpoint:      ExtendedEndpoint,
		AdminEndpoint: DeveloperEndpoint,
	},
	{
		Description:  "Create a workflow",
		Subject:      "workflow",
		AltSubject:   "wf",
		Predicate:    "create",
		AltPredicate: "new",
		FlagSet:      flag.NewFlagSet("create workflow", flag.ExitOnError),
		InitFunc: func(c *Command) {
			c.Arguments = map[string]interface{}{
				"label":               c.FlagSet.String("label", _nilDefaultStr, "Workflow's label."),
				"title":               c.FlagSet.String("title", _nilDefaultStr, "Workflow's title."),
				"usage":               c.FlagSet.String("usage", _nilDefaultStr, "Workflow's usage, one of:  infrastructure, network_equipment, server, free_standing, storage_pool, user, os_template."),
				"description":         c.FlagSet.String("description", _nilDefaultStr, "Workflow's description"),
				"deprecated":          c.FlagSet.Bool("deprecated", false, "Flag. Workflow's deprecation status. Default false"),
				"icon_asset_data_uri": c.FlagSet.String("icon", _nilDefaultStr, "Workflow's icon data"),
				"return_id":           c.FlagSet.Bool("return-id", false, green("(Flag)")+" If set will print the ID of the created workflow. Useful for automating tasks."),
			}
		},
		ExecuteFunc: workflowCreateCmd,
		Endpoint:    ExtendedEndpoint,
	},
	{
		Description:  "Delete a stage from a workflow.",
		Subject:      "workflow",
		AltSubject:   "wf",
		Predicate:    "delete-stage",
		AltPredicate: "rm-stage",
		FlagSet:      flag.NewFlagSet("delete workflow stage", flag.ExitOnError),
		InitFunc: func(c *Command) {
			c.Arguments = map[string]interface{}{
				"workflow_stage_id": c.FlagSet.Int("id", _nilDefaultInt, "Workflow's stage id "),
				"autoconfirm":       c.FlagSet.Bool("autoconfirm", false, green("(Flag)")+" If set it will assume action is confirmed"),
			}
		},
		ExecuteFunc:   workflowDeleteStageCmd,
		Endpoint:      ExtendedEndpoint,
		AdminEndpoint: DeveloperEndpoint,
	},
}

func workflowsListCmd(c *Command, client metalcloud.MetalCloudClient) (string, error) {

	usage := getStringParam(c.Arguments["usage"])

	list, err := client.WorkflowsWithUsage(usage)

	if err != nil {
		return "", err
	}

	schema := []tableformatter.SchemaField{
		{
			FieldName: "ID",
			FieldType: tableformatter.TypeInt,
			FieldSize: 6,
		},
		{
			FieldName: "LABEL",
			FieldType: tableformatter.TypeString,
			FieldSize: 6,
		},
		{
			FieldName: "USAGE",
			FieldType: tableformatter.TypeString,
			FieldSize: 5,
		},
		{
			FieldName: "DESCRIPTION",
			FieldType: tableformatter.TypeString,
			FieldSize: 5,
		},
		{
			FieldName: "TITLE",
			FieldType: tableformatter.TypeString,
			FieldSize: 5,
		},
		{
			FieldName: "OWNER",
			FieldType: tableformatter.TypeString,
			FieldSize: 5,
		},
		{
			FieldName: "DEPRECATED",
			FieldType: tableformatter.TypeBool,
			FieldSize: 5,
		},
		{
			FieldName: "CREATED",
			FieldType: tableformatter.TypeString,
			FieldSize: 5,
		},
		{
			FieldName: "UPDATED",
			FieldType: tableformatter.TypeString,
			FieldSize: 4,
		},
	}

	data := [][]interface{}{}
	for _, w := range *list {

		user := &metalcloud.User{
			UserID:          0,
			UserDisplayName: "",
			UserEmail:       "",
		}

		if w.UserIDOwner != 0 {
			user, err = client.UserGet(w.UserIDOwner)
			if err != nil {
				return "", err
			}
		}

		data = append(data, []interface{}{
			w.WorkflowID,
			w.WorkflowLabel,
			w.WorkflowUsage,
			w.WorkflowDescription,
			w.WorkflowTitle,
			user.UserEmail,
			w.WorkflowIsDeprecated,
			w.WorkflowCreatedTimestamp,
			w.WorkflowUpdatedTimestamp,
		})
	}

	tableformatter.TableSorter(schema).OrderBy(schema[0].FieldName).Sort(data)

	table := tableformatter.Table{
		Data:   data,
		Schema: schema,
	}
	return table.RenderTable("Workflows", "", getStringParam(c.Arguments["format"]))
}

func workflowGetCmd(c *Command, client metalcloud.MetalCloudClient) (string, error) {

	wf, err := getWorkflowFromCommand("id", c, client)
	if err != nil {
		return "", err
	}

	schema := []tableformatter.SchemaField{
		{
			FieldName: "RUNLEVEL",
			FieldType: tableformatter.TypeInt,
			FieldSize: 6,
		},
		{
			FieldName: "STAGES",
			FieldType: tableformatter.TypeString,
			FieldSize: 6,
		},
	}

	list, err := client.WorkflowStages(wf.WorkflowID)
	if err != nil {
		return "", err
	}

	runlevels := map[int][]string{}

	for _, s := range *list {
		stageDef, err := client.StageDefinitionGet(s.StageDefinitionID)
		if err != nil {
			return "", err
		}

		stageDescription := fmt.Sprintf("%s(#%d)-[WSI:# %d]",
			stageDef.StageDefinitionTitle,
			stageDef.StageDefinitionID,
			s.WorkflowStageID,
		)
		runlevels[s.WorkflowStageRunLevel] = append(runlevels[s.WorkflowStageRunLevel], stageDescription)
	}

	data := [][]interface{}{}
	for k, descriptions := range runlevels {

		data = append(data, []interface{}{
			k,
			strings.Join(descriptions, " "),
		})

	}

	tableformatter.TableSorter(schema).OrderBy(schema[0].FieldName).Sort(data)

	topLine := fmt.Sprintf("Workflow %s (%d) has the following stages:", wf.WorkflowLabel, wf.WorkflowID)
	table := tableformatter.Table{
		Data:   data,
		Schema: schema,
	}
	return table.RenderTable("Stages", topLine, getStringParam(c.Arguments["format"]))
}

func workflowCreateCmd(c *Command, client metalcloud.MetalCloudClient) (string, error) {

	label, ok := getStringParamOk(c.Arguments["label"])
	if !ok {
		return "", fmt.Errorf("-label is required")
	}

	usage, ok := getStringParamOk(c.Arguments["usage"])
	if !ok {
		return "", fmt.Errorf("-usage is required. It must be one of infrastructure, network_equipment, server, free_standing, storage_pool, user, os_template")
	}

	wf := metalcloud.Workflow{
		WorkflowLabel:        label,
		WorkflowTitle:        getStringParam(c.Arguments["title"]),
		WorkflowUsage:        usage,
		WorkflowDescription:  getStringParam(c.Arguments["description"]),
		WorkflowIsDeprecated: getBoolParam(c.Arguments["deprecated"]),
		IconAssetDataURI:     getStringParam(c.Arguments["icon"]),
	}

	ret, err := client.WorkflowCreate(wf)
	if err != nil {
		return "", err
	}
	if getBoolParam(c.Arguments["return_id"]) {
		return fmt.Sprintf("%d", ret.WorkflowID), nil
	}

	return "", nil

}

func workflowDeleteCmd(c *Command, client metalcloud.MetalCloudClient) (string, error) {

	ret, err := getWorkflowFromCommand("id", c, client)
	if err != nil {
		return "", err
	}
	confirm := false

	if getBoolParam(c.Arguments["autoconfirm"]) {
		confirm = true
	} else {

		confirmationMessage := fmt.Sprintf("Deleting workflow  %s (%d).  Are you sure? Type \"yes\" to continue:",
			ret.WorkflowTitle,
			ret.WorkflowID)

		//this is simply so that we don't output a text on the command line under go test
		if strings.HasSuffix(os.Args[0], ".test") {
			confirmationMessage = ""
		}

		confirm, err = requestConfirmation(confirmationMessage)
		if err != nil {
			return "", err
		}

	}

	if !confirm {
		return "", fmt.Errorf("Operation not confirmed. Aborting")
	}

	err = client.WorkflowDelete(ret.WorkflowID)

	return "", err
}

func workflowDeleteStageCmd(c *Command, client metalcloud.MetalCloudClient) (string, error) {

	workflowStageID, ok := getIntParamOk(c.Arguments["workflow_stage_id"])
	if !ok {
		return "", fmt.Errorf("-id is required (workflow-stage-id (WSI) number returned by get workflow")
	}

	workflowStage, err := client.WorkflowStageGet(workflowStageID)
	if err != nil {
		return "", err
	}

	confirm := getBoolParam(c.Arguments["autoconfirm"])

	if !confirm {

		wf, err := client.WorkflowGet(workflowStage.WorkflowID)
		if err != nil {
			return "", err
		}

		sd, err := client.StageDefinitionGet(workflowStage.StageDefinitionID)
		if err != nil {
			return "", err
		}

		confirmationMessage := fmt.Sprintf("Deleting stage %s (%d) from workflow %s (%d).  Are you sure? Type \"yes\" to continue:",
			wf.WorkflowTitle, wf.WorkflowID,
			sd.StageDefinitionTitle, sd.StageDefinitionID)

		//this is simply so that we don't output a text on the command line under go test
		if strings.HasSuffix(os.Args[0], ".test") {
			confirmationMessage = ""
		}

		confirm, err = requestConfirmation(confirmationMessage)
		if err != nil {
			return "", err
		}

	}

	if !confirm {
		return "", fmt.Errorf("Operation not confirmed. Aborting")
	}

	err = client.WorkflowStageDelete(workflowStageID)

	return "", err
}

func getWorkflowFromCommand(paramName string, c *Command, client metalcloud.MetalCloudClient) (*metalcloud.Workflow, error) {

	v, err := getParam(c, "workflow_id_or_label", paramName)
	if err != nil {
		return nil, err
	}

	id, label, isID := idOrLabel(v)

	if isID {
		return client.WorkflowGet(id)
	}

	list, err := client.Workflows()
	if err != nil {
		return nil, err
	}

	for _, s := range *list {
		if s.WorkflowLabel == label {
			return &s, nil
		}
	}

	if isID {
		return nil, fmt.Errorf("workflow %d not found", id)
	}

	return nil, fmt.Errorf("workflow %s not found", label)

}
