package volumetemplate

import (
	"flag"
	"fmt"
	"strings"

	metalcloud "github.com/metalsoft-io/metal-cloud-sdk-go/v3"
	"github.com/metalsoft-io/tableformatter"

	"github.com/metalsoft-io/metalcloud-cli/internal/colors"
	"github.com/metalsoft-io/metalcloud-cli/internal/command"
	"github.com/metalsoft-io/metalcloud-cli/internal/configuration"
)

var VolumeTemplateCmds = []command.Command{
	{
		Description:  "Lists available volume templates.",
		Subject:      "volume-template",
		AltSubject:   "vt",
		Predicate:    "list",
		AltPredicate: "ls",
		FlagSet:      flag.NewFlagSet("list volume templates", flag.ExitOnError),
		InitFunc: func(c *command.Command) {
			c.Arguments = map[string]interface{}{
				"format":     c.FlagSet.String("format", command.NilDefaultStr, "The output format. Supported values are 'json','csv','yaml'. The default format is human readable."),
				"local_only": c.FlagSet.Bool("local-only", false, "Show only templates that support local install"),
				"pxe_only":   c.FlagSet.Bool("pxe-only", false, "Show only templates that support pxe booting"),
			}
		},
		ExecuteFunc:         volumeTemplatesListCmd,
		Endpoint:            configuration.ExtendedEndpoint,
		PermissionsRequired: []string{command.TEMPLATES_READ}, //while the user would have access to the volume template are deprecating this
	},
	{
		Description:  "Create volume templates.",
		Subject:      "volume-template",
		AltSubject:   "vt",
		Predicate:    "create",
		AltPredicate: "new",
		FlagSet:      flag.NewFlagSet("create volume templates", flag.ExitOnError),
		InitFunc: func(c *command.Command) {
			c.Arguments = map[string]interface{}{
				"drive_id":                   c.FlagSet.Int("id", command.NilDefaultInt, colors.Red("(Required)")+" The id of the drive to create the volume template from"),
				"label":                      c.FlagSet.String("label", command.NilDefaultStr, colors.Red("(Required)")+" The label of the volume template"),
				"description":                c.FlagSet.String("description", command.NilDefaultStr, colors.Red("(Required)")+" The description of the volume template"),
				"display_name":               c.FlagSet.String("name", command.NilDefaultStr, colors.Red("(Required)")+" The display name of the volume template"),
				"boot_type":                  c.FlagSet.String("boot-type", command.NilDefaultStr, "The boot_type of the volume template. Possible values: 'uefi_only','legacy_only' "),
				"boot_methods_supported":     c.FlagSet.String("boot-methods-supported", command.NilDefaultStr, "The boot_methods_supported of the volume template. Defaults to 'pxe_iscsi'."),
				"deprecation_status":         c.FlagSet.String("deprecation-status", command.NilDefaultStr, "Deprecation status. Possible values: not_deprecated,deprecated_deny_provision,deprecated_allow_expand. Defaults to 'not_deprecated'."),
				"tags":                       c.FlagSet.String("tags", command.NilDefaultStr, "The tags of the volume template, comma separated."),
				"os_bootstrap_function_name": c.FlagSet.String("os-bootstrap-function-name", command.NilDefaultStr, "Optional property that selects the cloudinit configuration function. Can be one of: provisioner_os_cloudinit_prepare_centos, provisioner_os_cloudinit_prepare_rhel, provisioner_os_cloudinit_prepare_ubuntu, provisioner_os_cloudinit_prepare_windows."),
				"os_type":                    c.FlagSet.String("os-type", command.NilDefaultStr, "Template operating system type. For example, Ubuntu or CentOS. If set, os-version and os-architecture flags are required as well."),
				"os_version":                 c.FlagSet.String("os-version", command.NilDefaultStr, "Template operating system version. If set, os-type and os-architecture flags are required as well."),
				"os_architecture":            c.FlagSet.String("os-architecture", command.NilDefaultStr, "Template operating system architecture.Possible values: none, unknown, x86, x86_64. If set, os-version and os-type flags are required as well."),
				"version":                    c.FlagSet.String("version", command.NilDefaultStr, "Template version. Default value is 0.0.0"),
				"os_ready_method":            c.FlagSet.String("os-ready-method", command.NilDefaultStr, "Possible values: 'wait_for_ssh', 'wait_for_signal_from_os'. Default value: 'wait_for_ssh'."),
				"network_os":                 c.FlagSet.Bool("network-os", false, colors.Green("(Flag)")+" Must be set for network operating system (NOS) templates."),
				"network_os_switch_driver":   c.FlagSet.String("network-os-switch-driver", command.NilDefaultStr, colors.Yellow("(Required if network_os)")+"Network operating system (NOS) switch driver, e.g. 'sonic_enterprise'."),
				"network_os_switch_role":     c.FlagSet.String("network-os-switch-role", command.NilDefaultStr, colors.Yellow("(Flag)")+"Network operating system (NOS) switch role. Possible values: 'leaf', 'spine'."),
				"network_os_datacenter_name": c.FlagSet.String("network-os-datacenter-name", command.NilDefaultStr, colors.Yellow("(Flag)")+"Network operating system (NOS) datacenter name, e.g. 'dc1'"),
				"network_os_version":         c.FlagSet.String("network-os-version", command.NilDefaultStr, colors.Yellow("(Required if network_os)")+"Network operating system (NOS) version, e.g. '4.0.2'"),
				"network_os_architecture":    c.FlagSet.String("network-os-architecture", command.NilDefaultStr, colors.Yellow("(Required if network_os)")+"Network operating system (NOS) architecture. Possible values: 'x86_64', 'aarch64'."),
				"network_os_vendor":          c.FlagSet.String("network-os-vendor", command.NilDefaultStr, colors.Yellow("(Required if network_os)")+"Network operating system (NOS) vendor, e.g. 'dellemc'"),
				"network_os_machine":         c.FlagSet.String("network-os-machine", command.NilDefaultStr, colors.Yellow("(Required if network_os)")+"Network operating system (NOS) machine(equipment model) e.g.'s5212f_c3538'"),
				"return_id":                  c.FlagSet.Bool("return-id", false, "(Optional) Will print the ID of the created Volume Template. Useful for automating tasks."),
			}
		},
		ExecuteFunc:         volumeTemplateCreateFromDriveCmd,
		Endpoint:            configuration.ExtendedEndpoint,
		PermissionsRequired: []string{command.TEMPLATES_WRITE}, //while the user would have access to the volume template are deprecating this
	},
	{
		Description:  "Allow other users of the platform to use the template.",
		Subject:      "volume-template",
		AltSubject:   "vt",
		Predicate:    "make-public",
		AltPredicate: "public",
		FlagSet:      flag.NewFlagSet("make volume template public", flag.ExitOnError),
		InitFunc: func(c *command.Command) {
			c.Arguments = map[string]interface{}{
				"template_id_or_name":        c.FlagSet.String("id", command.NilDefaultStr, "Volume template id or name"),
				"os_bootstrap_function_name": c.FlagSet.String("os-bootstrap-function-name", command.NilDefaultStr, colors.Red("(Required)")+" Selects the cloudinit configuration function. Can be one of: provisioner_os_cloudinit_prepare_centos, provisioner_os_cloudinit_prepare_rhel, provisioner_os_cloudinit_prepare_ubuntu, provisioner_os_cloudinit_prepare_windows."),
			}
		},
		ExecuteFunc:         volumeTemplateMakePublicCmd,
		Endpoint:            configuration.DeveloperEndpoint,
		PermissionsRequired: []string{command.TEMPLATES_WRITE}, //while the user would have access to the volume template are deprecating this
	},
	{
		Description:  "Stop other users of the platform from being able to use the template by allocating a specific owner.",
		Subject:      "volume-template",
		AltSubject:   "vt",
		Predicate:    "make-private",
		AltPredicate: "private",
		FlagSet:      flag.NewFlagSet("make volume template private", flag.ExitOnError),
		InitFunc: func(c *command.Command) {
			c.Arguments = map[string]interface{}{
				"template_id_or_name": c.FlagSet.String("id", command.NilDefaultStr, "Volume template id or name"),
				"user_id":             c.FlagSet.String("user-id", command.NilDefaultStr, "New owner user id or email."),
			}
		},
		ExecuteFunc:         volumeTemplateMakePrivateCmd,
		Endpoint:            configuration.DeveloperEndpoint,
		PermissionsRequired: []string{command.TEMPLATES_WRITE}, //while the user would have access to the volume template are deprecating this
	},
}

func volumeTemplatesListCmd(c *command.Command, client metalcloud.MetalCloudClient) (string, error) {

	vList, err := client.VolumeTemplates()
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
			FieldSize: 20,
		},
		{
			FieldName: "NAME",
			FieldType: tableformatter.TypeString,
			FieldSize: 20,
		},
		{
			FieldName: "SIZE",
			FieldType: tableformatter.TypeInt,
			FieldSize: 6,
		},
		{
			FieldName: "STATUS",
			FieldType: tableformatter.TypeString,
			FieldSize: 20,
		},
		{
			FieldName: "FLAGS",
			FieldType: tableformatter.TypeString,
			FieldSize: 10,
		},
	}

	localOnly := c.Arguments["local_only"] != nil && *c.Arguments["local_only"].(*bool)
	pxeOnly := c.Arguments["pxe_only"] != nil && *c.Arguments["pxe_only"].(*bool)

	data := [][]interface{}{}
	for _, v := range *vList {

		if localOnly && !v.VolumeTemplateLocalDiskSupported {
			continue
		}

		if pxeOnly && !strings.Contains(v.VolumeTemplateBootMethodsSupported, "pxe_iscsi") {
			continue
		}

		flags := []string{}

		flags = append(flags, strings.Split(v.VolumeTemplateBootMethodsSupported, ",")...)

		if v.VolumeTemplateLocalDiskSupported {
			flags = append(flags, "local")
		}

		data = append(data, []interface{}{
			v.VolumeTemplateID,
			v.VolumeTemplateLabel,
			v.VolumeTemplateDisplayName,
			v.VolumeTemplateSizeMBytes,
			v.VolumeTemplateDeprecationStatus,
			strings.Join(flags, ","),
		})

	}

	tableformatter.TableSorter(schema).OrderBy(schema[0].FieldName).Sort(data)

	table := tableformatter.Table{
		Data:   data,
		Schema: schema,
	}
	return table.RenderTable("Volume templates", "", command.GetStringParam(c.Arguments["format"]))
}

func volumeTemplateCreateFromDriveCmd(c *command.Command, client metalcloud.MetalCloudClient) (string, error) {
	objVolumeTemplate := metalcloud.VolumeTemplate{
		VolumeTemplateDeprecationStatus:       command.GetStringParam(c.Arguments["deprecation_status"]),
		VolumeTemplateBootMethodsSupported:    command.GetStringParam(c.Arguments["boot_methods_supported"]),
		VolumeTemplateTags:                    strings.Split(command.GetStringParam(c.Arguments["tags"]), ","),
		VolumeTemplateBootType:                command.GetStringParam(c.Arguments["boot_type"]),
		VolumeTemplateOsBootstrapFunctionName: command.GetStringParam(c.Arguments["os_bootstrap_function_name"]),
		VolumeTemplateVersion:                 command.GetStringParam(c.Arguments["version"]),
		VolumeTemplateOSReadyMethod:           command.GetStringParam(c.Arguments["os_ready_method"]),
	}

	driveID, ok := command.GetIntParamOk(c.Arguments["drive_id"])
	if !ok {
		return "", fmt.Errorf("-id is required (drive id)")
	}

	if label, ok := command.GetStringParamOk(c.Arguments["label"]); ok {
		objVolumeTemplate.VolumeTemplateLabel = label
	} else {
		return "", fmt.Errorf("-label is required")
	}

	if description, ok := command.GetStringParamOk(c.Arguments["description"]); ok {
		objVolumeTemplate.VolumeTemplateDescription = description
	} else {
		return "", fmt.Errorf("-description is required")
	}

	if name, ok := command.GetStringParamOk(c.Arguments["display_name"]); ok {
		objVolumeTemplate.VolumeTemplateDisplayName = name
	} else {
		return "", fmt.Errorf("-name is required")
	}

	os, err := getOperatingSystemFromCommand(c)

	if err != nil {
		return "", err
	}
	objVolumeTemplate.VolumeTemplateOperatingSystem = *os

	ret, err := client.VolumeTemplateCreateFromDrive(driveID, objVolumeTemplate)
	if err != nil {
		return "", err
	}

	if command.GetBoolParam(c.Arguments["return_id"]) {
		return fmt.Sprintf("%d", ret.VolumeTemplateID), nil
	}

	return "", nil
}

func getOperatingSystemFromCommand(c *command.Command) (*metalcloud.OperatingSystem, error) {
	var operatingSystem = metalcloud.OperatingSystem{}
	present := false

	if osType, ok := command.GetStringParamOk(c.Arguments["os_type"]); ok {
		present = true
		operatingSystem.OperatingSystemType = osType
	}

	if osVersion, ok := command.GetStringParamOk(c.Arguments["os_version"]); ok {
		if !present {
			return nil, fmt.Errorf("some of the operating system flags are missing")
		}
		operatingSystem.OperatingSystemVersion = osVersion
	} else if present {
		return nil, fmt.Errorf("os-version is required")
	}

	if osArchitecture, ok := command.GetStringParamOk(c.Arguments["os_architecture"]); ok {
		if !present {
			return nil, fmt.Errorf("some of the operating system flags are missing")
		}
		operatingSystem.OperatingSystemArchitecture = osArchitecture
	} else if present {
		return nil, fmt.Errorf("os-architecture is required")
	}

	return &operatingSystem, nil
}

func getNetworkOperatingSystemFromCommand(c *command.Command) (*metalcloud.NetworkOperatingSystem, error) {
	var networkOperatingSystem = metalcloud.NetworkOperatingSystem{}

	nosSwitchDriver := command.GetStringParam(c.Arguments["network_os_switch_driver"])
	if nosSwitchDriver != "" {
		networkOperatingSystem.OperatingSystemSwitchDriver = nosSwitchDriver
	} else {
		return nil, fmt.Errorf("network-os-switch-driver is required")
	}

	nosSwitchRole := command.GetStringParam(c.Arguments["network_os_switch_role"])
	if nosSwitchRole != "" {
		networkOperatingSystem.OperatingSystemSwitchRole = nosSwitchRole
	}

	nosVersion := command.GetStringParam(c.Arguments["network_os_version"])
	if nosVersion != "" {
		networkOperatingSystem.OperatingSystemVersion = nosVersion
	} else {
		return nil, fmt.Errorf("network-os-version is required")
	}

	nosArchitecture := command.GetStringParam(c.Arguments["network_os_architecture"])
	if nosArchitecture != "" {
		networkOperatingSystem.OperatingSystemArchitecture = nosArchitecture
	} else {
		return nil, fmt.Errorf("network-os-architecture is required")
	}

	nosVendor := command.GetStringParam(c.Arguments["network_os_vendor"])
	if nosVendor != "" {
		networkOperatingSystem.OperatingSystemVendor = nosVendor
	} else {
		return nil, fmt.Errorf("network-os-vendor is required")
	}

	nosMachine := command.GetStringParam(c.Arguments["network_os_machine"])
	if nosMachine != "" {
		networkOperatingSystem.OperatingSystemMachine = nosMachine
	} else {
		return nil, fmt.Errorf("network-os-machine is required")
	}

	nosDatacenterName := command.GetStringParam(c.Arguments["network_os_datacenter_name"])
	if nosDatacenterName != "" {
		networkOperatingSystem.OperatingSystemDatacenterName = nosDatacenterName
	}

	return &networkOperatingSystem, nil
}

func getVolumeTemplateFromCommand(paramName string, c *command.Command, client metalcloud.MetalCloudClient) (*metalcloud.VolumeTemplate, error) {
	v, err := command.GetParam(c, "template_id_or_name", paramName)
	if err != nil {
		return nil, err
	}

	id, label, isID := command.IdOrLabel(v)

	if isID {
		return client.VolumeTemplateGet(id)
	}

	list, err := client.VolumeTemplates()
	if err != nil {
		return nil, err
	}

	for _, s := range *list {
		if s.VolumeTemplateLabel == label {
			return &s, nil
		}
	}

	if isID {
		return nil, fmt.Errorf("volume template %d not found", id)
	}

	return nil, fmt.Errorf("volume template %s not found", label)
}

func volumeTemplateMakePublicCmd(c *command.Command, client metalcloud.MetalCloudClient) (string, error) {
	template, err := getVolumeTemplateFromCommand("id", c, client)

	if err != nil {
		return "", err
	}

	osBootstrapFunctionName, ok := command.GetStringParamOk(c.Arguments["os_bootstrap_function_name"])
	if !ok {
		return "", fmt.Errorf("-os-bootstrap-function-name is required")
	}

	err = client.VolumeTemplateMakePublic(template.VolumeTemplateID, osBootstrapFunctionName)

	if err != nil {
		return "", err
	}

	return "", nil
}

func volumeTemplateMakePrivateCmd(c *command.Command, client metalcloud.MetalCloudClient) (string, error) {
	template, err := getVolumeTemplateFromCommand("id", c, client)

	if err != nil {
		return "", err
	}

	user, err := command.GetUserFromCommand("user-id", c, client)
	if err != nil {
		return "", err
	}

	if err = client.VolumeTemplateMakePrivate(template.VolumeTemplateID, user.UserID); err != nil {
		return "", err
	}

	return "", nil
}
