package firewall

import (
	"flag"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	metalcloud "github.com/metalsoft-io/metal-cloud-sdk-go/v3"
	"github.com/metalsoft-io/metalcloud-cli/internal/colors"
	"github.com/metalsoft-io/metalcloud-cli/internal/command"
	"github.com/metalsoft-io/tableformatter"
)

var FirewallRuleCmds = []command.Command{
	{
		Description:  "Lists instance array firewall rules.",
		Subject:      "firewall-rule",
		AltSubject:   "fw",
		Predicate:    "list",
		AltPredicate: "ls",
		FlagSet:      flag.NewFlagSet("list firewall rules", flag.ExitOnError),
		InitFunc: func(c *command.Command) {
			c.Arguments = map[string]interface{}{
				"instance_array_id": c.FlagSet.Int("ia", command.NilDefaultInt, colors.Red("(Required)")+" The instance array id"),
				"format":            c.FlagSet.String("format", command.NilDefaultStr, "The output format. Supported values are 'json','csv','yaml'. The default format is human readable."),
			}
		},
		ExecuteFunc: firewallRuleListCmd,
	},
	{
		Description:  "Add instance array firewall rule.",
		Subject:      "firewall-rule",
		AltSubject:   "fw",
		Predicate:    "add",
		AltPredicate: "new",
		FlagSet:      flag.NewFlagSet("add firewall rules", flag.ExitOnError),
		InitFunc: func(c *command.Command) {
			c.Arguments = map[string]interface{}{
				"instance_array_id":                   c.FlagSet.Int("ia", command.NilDefaultInt, colors.Red("(Required)")+" The instance array id"),
				"firewall_rule_protocol":              c.FlagSet.String("protocol", command.NilDefaultStr, "The protocol of the firewall rule. Possible values: all, icmp, tcp, udp."),
				"firewall_rule_ip_address_type":       c.FlagSet.String("ip-address-type", "ipv4", "The IP address type of the firewall rule. Possible values: ipv4, ipv6."),
				"firewall_rule_port":                  c.FlagSet.String("port", command.NilDefaultStr, "The port to filter on. It can also be a range with the start and end values separated by a dash."),
				"firewall_rule_source_ip_address":     c.FlagSet.String("source", command.NilDefaultStr, "The source address to filter on. It can also be a range with the start and end values separated by a dash."),
				"firewall_rule_desination_ip_address": c.FlagSet.String("destination", command.NilDefaultStr, "The destination address to filter on. It can also be a range with the start and end values separated by a dash."),
				"firewall_rule_description":           c.FlagSet.String("description", command.NilDefaultStr, "The firewall rule's description."),
			}
		},
		ExecuteFunc: firewallRuleAddCmd,
	},
	{
		Description:  "Remove instance array firewall rule.",
		Subject:      "firewall-rule",
		AltSubject:   "fw",
		Predicate:    "delete",
		AltPredicate: "rm",
		FlagSet:      flag.NewFlagSet("delete firewall rules", flag.ExitOnError),
		InitFunc: func(c *command.Command) {
			c.Arguments = map[string]interface{}{
				"instance_array_id":                   c.FlagSet.Int("ia", command.NilDefaultInt, colors.Red("(Required)")+" The instance array id"),
				"firewall_rule_ip_address_type":       c.FlagSet.String("ip-address-type", "ipv4", "The IP address type of the firewall rule. Possible values: ipv4, ipv6."),
				"firewall_rule_protocol":              c.FlagSet.String("protocol", command.NilDefaultStr, "The protocol of the firewall rule. Possible values: all, icmp, tcp, udp."),
				"firewall_rule_port":                  c.FlagSet.String("port", command.NilDefaultStr, "The port to filter on. It can also be a range with the start and end values separated by a dash."),
				"firewall_rule_source_ip_address":     c.FlagSet.String("source", command.NilDefaultStr, "The source address to filter on. It can also be a range with the start and end values separated by a dash."),
				"firewall_rule_desination_ip_address": c.FlagSet.String("destination", command.NilDefaultStr, "The destination address to filter on. It can also be a range with the start and end values separated by a dash."),
			}
		},
		ExecuteFunc: firewallRuleDeleteCmd,
	},
}

func firewallRuleListCmd(c *command.Command, client metalcloud.MetalCloudClient) (string, error) {

	instanceArrayID := c.Arguments["instance_array_id"]

	if instanceArrayID == nil || *instanceArrayID.(*int) == 0 {
		return "", fmt.Errorf("-ia is required")
	}

	retIA, err := client.InstanceArrayGet(*instanceArrayID.(*int))
	if err != nil {
		return "", err
	}

	if !retIA.InstanceArrayOperation.InstanceArrayFirewallManaged {
		return "", fmt.Errorf("the instance array %s [#%d] has firewall management disabled", retIA.InstanceArrayLabel, retIA.InstanceArrayID)
	}

	schema := []tableformatter.SchemaField{
		{
			FieldName: "INDEX",
			FieldType: tableformatter.TypeInt,
			FieldSize: 6,
		},
		{
			FieldName: "PROTOCOL",
			FieldType: tableformatter.TypeString,
			FieldSize: 10,
		},
		{
			FieldName: "PORT",
			FieldType: tableformatter.TypeString,
			FieldSize: 10,
		},
		{
			FieldName: "SOURCE",
			FieldType: tableformatter.TypeString,
			FieldSize: 20,
		},
		{
			FieldName: "DEST",
			FieldType: tableformatter.TypeString,
			FieldSize: 20,
		},
		{
			FieldName: "TYPE",
			FieldType: tableformatter.TypeString,
			FieldSize: 5,
		},
		{
			FieldName: "ENABLED",
			FieldType: tableformatter.TypeBool,
			FieldSize: 10,
		},
		{
			FieldName: "DESC.",
			FieldType: tableformatter.TypeString,
			FieldSize: 50,
		},
	}

	status := retIA.InstanceArrayServiceStatus
	if retIA.InstanceArrayServiceStatus != "ordered" && retIA.InstanceArrayOperation.InstanceArrayDeployType == "edit" && retIA.InstanceArrayOperation.InstanceArrayDeployStatus == "not_started" {
		status = "edited"
	}

	list := retIA.InstanceArrayOperation.InstanceArrayFirewallRules
	data := [][]interface{}{}
	idx := 0
	for _, fw := range list {

		portRange := "any"

		if fw.FirewallRulePortRangeStart != 0 {
			portRange = fmt.Sprintf("%d", fw.FirewallRulePortRangeStart)
		}

		if fw.FirewallRulePortRangeStart != fw.FirewallRulePortRangeEnd {
			portRange += fmt.Sprintf("-%d", fw.FirewallRulePortRangeEnd)
		}

		sourceIPRange := "any"

		if fw.FirewallRuleSourceIPAddressRangeStart != "" {
			sourceIPRange = fw.FirewallRuleSourceIPAddressRangeStart
		}

		if fw.FirewallRuleSourceIPAddressRangeStart != fw.FirewallRuleSourceIPAddressRangeEnd {
			sourceIPRange += fmt.Sprintf("-%s", fw.FirewallRuleSourceIPAddressRangeEnd)
		}

		destinationIPRange := "any"

		if fw.FirewallRuleDestinationIPAddressRangeStart != "" {
			sourceIPRange = fw.FirewallRuleSourceIPAddressRangeStart
		}

		if fw.FirewallRuleDestinationIPAddressRangeStart != fw.FirewallRuleDestinationIPAddressRangeEnd {
			sourceIPRange += fmt.Sprintf("-%s", fw.FirewallRuleDestinationIPAddressRangeEnd)
		}

		data = append(data, []interface{}{
			idx,
			fw.FirewallRuleProtocol,
			portRange,
			sourceIPRange,
			destinationIPRange,
			fw.FirewallRuleIPAddressType,
			fw.FirewallRuleEnabled,
			fw.FirewallRuleDescription,
		})

		idx++

	}

	topLine := fmt.Sprintf("Instance Array %s (%d) [%s] has the following firewall rules:\n", retIA.InstanceArrayLabel, retIA.InstanceArrayID, status)

	tableformatter.TableSorter(schema).OrderBy(schema[0].FieldName).Sort(data)

	table := tableformatter.Table{
		Data:   data,
		Schema: schema,
	}
	return table.RenderTable("Rules", topLine, command.GetStringParam(c.Arguments["format"]))
}

func firewallRuleAddCmd(c *command.Command, client metalcloud.MetalCloudClient) (string, error) {
	instanceArrayID := c.Arguments["instance_array_id"]

	if instanceArrayID == nil || *instanceArrayID.(*int) == 0 {
		return "", fmt.Errorf("-ia is required")
	}

	retIA, err := client.InstanceArrayGet(*instanceArrayID.(*int))
	if err != nil {
		return "", err
	}

	fw := metalcloud.FirewallRule{}

	if v := c.Arguments["firewall_rule_protocol"]; v != nil && *v.(*string) != command.NilDefaultStr {
		fw.FirewallRuleProtocol = *v.(*string)
	}

	if v := c.Arguments["firewall_rule_ip_address_type"]; v != nil && *v.(*string) != command.NilDefaultStr {
		fw.FirewallRuleIPAddressType = *v.(*string)
	}

	if v := c.Arguments["firewall_rule_port"]; v != nil && *v.(*string) != command.NilDefaultStr {
		fw.FirewallRulePortRangeStart, fw.FirewallRulePortRangeEnd, err = portStringToRange(*v.(*string))
		if err != nil {
			return "", err
		}
	}

	if v := c.Arguments["firewall_rule_source_ip_address"]; v != nil && *v.(*string) != command.NilDefaultStr {
		fw.FirewallRuleSourceIPAddressRangeStart, fw.FirewallRuleSourceIPAddressRangeEnd, err = addressStringToRange(*v.(*string))
		if err != nil {
			return "", err
		}
	}

	if v := c.Arguments["firewall_rule_desination_ip_address"]; v != nil && *v.(*string) != command.NilDefaultStr {
		fw.FirewallRuleDestinationIPAddressRangeStart, fw.FirewallRuleDestinationIPAddressRangeEnd, err = addressStringToRange(*v.(*string))
		if err != nil {
			return "", err
		}
	}

	if v := c.Arguments["firewall_rule_description"]; v != nil && *v.(*string) != command.NilDefaultStr {
		fw.FirewallRuleDescription = *v.(*string)
	}

	retIA.InstanceArrayOperation.InstanceArrayFirewallRules = append(
		retIA.InstanceArrayOperation.InstanceArrayFirewallRules,
		fw)

	bFalse := false
	_, err = client.InstanceArrayEdit(retIA.InstanceArrayID, *retIA.InstanceArrayOperation, &bFalse, nil, nil, nil)

	return "", err
}

func firewallRuleDeleteCmd(c *command.Command, client metalcloud.MetalCloudClient) (string, error) {
	instanceArrayID := c.Arguments["instance_array_id"]

	if instanceArrayID == nil || *instanceArrayID.(*int) == 0 {
		return "", fmt.Errorf("-ia is required")
	}

	retIA, err := client.InstanceArrayGet(*instanceArrayID.(*int))
	if err != nil {
		return "", err
	}

	fw := metalcloud.FirewallRule{}

	if v := c.Arguments["firewall_rule_protocol"]; v != nil && *v.(*string) != command.NilDefaultStr {
		fw.FirewallRuleProtocol = *v.(*string)
	}

	if v := c.Arguments["firewall_rule_ip_address_type"]; v != nil && *v.(*string) != command.NilDefaultStr {
		fw.FirewallRuleIPAddressType = *v.(*string)
	}

	if v := c.Arguments["firewall_rule_port"]; v != nil && *v.(*string) != command.NilDefaultStr {
		fw.FirewallRulePortRangeStart, fw.FirewallRulePortRangeEnd, err = portStringToRange(*v.(*string))
		if err != nil {
			return "", err
		}
	}

	if v := c.Arguments["firewall_rule_source_ip_address"]; v != nil && *v.(*string) != command.NilDefaultStr {
		fw.FirewallRuleSourceIPAddressRangeStart, fw.FirewallRuleSourceIPAddressRangeEnd, err = addressStringToRange(*v.(*string))
		if err != nil {
			return "", err
		}
	}

	if v := c.Arguments["firewall_rule_desination_ip_address"]; v != nil && *v.(*string) != command.NilDefaultStr {
		fw.FirewallRuleDestinationIPAddressRangeStart, fw.FirewallRuleDestinationIPAddressRangeEnd, err = addressStringToRange(*v.(*string))
		if err != nil {
			return "", err
		}
	}

	newFW := []metalcloud.FirewallRule{}
	found := false
	for _, f := range retIA.InstanceArrayOperation.InstanceArrayFirewallRules {
		if !fwRulesEqual(f, fw) {
			newFW = append(newFW, f)
		} else {
			found = true
		}
	}

	if !found {
		return "", fmt.Errorf("No matching firewall rule was found %v", fw)
	}

	retIA.InstanceArrayOperation.InstanceArrayFirewallRules = newFW
	bFalse := false
	_, err = client.InstanceArrayEdit(retIA.InstanceArrayID, *retIA.InstanceArrayOperation, &bFalse, nil, nil, nil)

	return "", err
}

func fwRulesEqual(a, b metalcloud.FirewallRule) bool {
	return a.FirewallRuleProtocol == b.FirewallRuleProtocol &&
		a.FirewallRulePortRangeStart == b.FirewallRulePortRangeStart &&
		a.FirewallRulePortRangeEnd == b.FirewallRulePortRangeEnd &&
		a.FirewallRuleSourceIPAddressRangeStart == b.FirewallRuleSourceIPAddressRangeStart &&
		a.FirewallRuleSourceIPAddressRangeEnd == b.FirewallRuleSourceIPAddressRangeEnd &&
		a.FirewallRuleDestinationIPAddressRangeStart == b.FirewallRuleDestinationIPAddressRangeStart &&
		a.FirewallRuleDestinationIPAddressRangeEnd == b.FirewallRuleDestinationIPAddressRangeEnd
}

func portStringToRange(s string) (int, int, error) {
	port, err := strconv.Atoi(s)

	if err == nil && port > 0 {
		return port, port, nil
	}

	re := regexp.MustCompile(`^(\d+)\-(\d+)$`)
	matches := re.FindStringSubmatch(s)

	if matches == nil {
		return 0, 0, fmt.Errorf("Could not parse port definition %s", s)
	}

	startPort, err := strconv.Atoi(matches[1])

	if err != nil && startPort > 0 {
		return 0, 0, fmt.Errorf("Could not parse port definition %s", s)
	}

	endPort, err := strconv.Atoi(matches[2])
	if err != nil && endPort > 0 {
		return 0, 0, fmt.Errorf("Could not parse port definition %s", s)
	}

	return startPort, endPort, nil

}

func addressStringToRange(s string) (string, string, error) {

	if s == "" {
		return "", "", fmt.Errorf("address cannot be empty")
	}

	components := strings.Split(s, "-")

	if len(components) == 1 && components[0] != "" {
		return s, s, nil //single address, we return it
	}

	if len(components) != 2 || components[0] == "" || components[1] == "" {
		return "", "", fmt.Errorf("cannot parse address %s", s)
	}

	return components[0], components[1], nil

}
