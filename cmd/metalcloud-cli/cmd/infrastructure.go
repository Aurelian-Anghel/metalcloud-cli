package cmd

import (
	"github.com/metalsoft-io/metalcloud-cli/cmd/metalcloud-cli/system"
	"github.com/metalsoft-io/metalcloud-cli/internal/infrastructure"
	"github.com/spf13/cobra"
)

var (
	showAll             bool
	showOrdered         bool
	showDeleted         bool
	customVariables     string
	allowDataLoss       bool
	attemptSoftShutdown bool
	attemptHardShutdown bool
	softShutdownTimeout int
	forceShutdown       bool

	infrastructureCmd = &cobra.Command{
		Use:     "infrastructure [command]",
		Aliases: []string{"infra"},
		Short:   "Infrastructure management",
		Long:    `Infrastructure management commands.`,
	}

	infrastructureListCmd = &cobra.Command{
		Use:          "list [flags...]",
		Aliases:      []string{"ls"},
		Short:        "List all infrastructures.",
		SilenceUsage: true,
		Annotations:  map[string]string{system.REQUIRED_PERMISSION: system.SITE_READ}, // TODO: Use specific permission
		RunE: func(cmd *cobra.Command, args []string) error {
			return infrastructure.InfrastructureList(cmd.Context(), showAll, showOrdered, showDeleted)
		},
	}

	infrastructureGetCmd = &cobra.Command{
		Use:          "get infrastructure_id_or_label",
		Aliases:      []string{"show"},
		Short:        "Get infrastructure details.",
		SilenceUsage: true,
		Annotations:  map[string]string{system.REQUIRED_PERMISSION: system.SITE_READ}, // TODO: Use specific permission
		Args:         cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return infrastructure.InfrastructureGet(cmd.Context(), args[0])
		},
	}

	infrastructureCreateCmd = &cobra.Command{
		Use:          "create site_id label",
		Aliases:      []string{"new"},
		Short:        "Create new infrastructure.",
		SilenceUsage: true,
		Annotations:  map[string]string{system.REQUIRED_PERMISSION: system.SITE_WRITE}, // TODO: Use specific permission
		Args:         cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return infrastructure.InfrastructureCreate(cmd.Context(), args[0], args[1])
		},
	}

	infrastructureUpdateCmd = &cobra.Command{
		Use:          "update infrastructure_id_or_label [new_label]",
		Aliases:      []string{"edit"},
		Short:        "Update infrastructure configuration.",
		SilenceUsage: true,
		Annotations:  map[string]string{system.REQUIRED_PERMISSION: system.SITE_WRITE}, // TODO: Use specific permission
		Args:         cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			label := ""
			if len(args) > 1 {
				label = args[1]
			}

			return infrastructure.InfrastructureUpdate(cmd.Context(), args[0], label, customVariables)
		},
	}

	infrastructureDeleteCmd = &cobra.Command{
		Use:          "delete infrastructure_id_or_label",
		Aliases:      []string{"rm"},
		Short:        "Delete infrastructure.",
		SilenceUsage: true,
		Annotations:  map[string]string{system.REQUIRED_PERMISSION: system.SITE_WRITE}, // TODO: Use specific permission
		Args:         cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return infrastructure.InfrastructureDelete(cmd.Context(), args[0])
		},
	}

	infrastructureDeployCmd = &cobra.Command{
		Use:          "deploy infrastructure_id_or_label",
		Aliases:      []string{"apply"},
		Short:        "Deploy infrastructure.",
		SilenceUsage: true,
		Annotations:  map[string]string{system.REQUIRED_PERMISSION: system.SITE_WRITE}, // TODO: Use specific permission
		Args:         cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return infrastructure.InfrastructureDeploy(cmd.Context(), args[0], allowDataLoss, attemptSoftShutdown, attemptHardShutdown, softShutdownTimeout, forceShutdown)
		},
	}

	infrastructureRevertCmd = &cobra.Command{
		Use:          "revert infrastructure_id_or_label",
		Aliases:      []string{"undo"},
		Short:        "Revert infrastructure changes.",
		SilenceUsage: true,
		Annotations:  map[string]string{system.REQUIRED_PERMISSION: system.SITE_WRITE}, // TODO: Use specific permission
		Args:         cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return infrastructure.InfrastructureRevert(cmd.Context(), args[0])
		},
	}
)

func init() {
	rootCmd.AddCommand(infrastructureCmd)

	infrastructureCmd.AddCommand(infrastructureListCmd)
	infrastructureListCmd.Flags().BoolVar(&showAll, "show-all", false, "If set will return all infrastructures.")
	infrastructureListCmd.Flags().BoolVar(&showOrdered, "show-ordered", false, "If set will also return ordered (created but not deployed) infrastructures.")
	infrastructureListCmd.Flags().BoolVar(&showDeleted, "show-deleted", false, "If set will also return deleted infrastructures.")

	infrastructureCmd.AddCommand(infrastructureGetCmd)

	infrastructureCmd.AddCommand(infrastructureCreateCmd)

	infrastructureCmd.AddCommand(infrastructureUpdateCmd)
	infrastructureUpdateCmd.Flags().StringVar(&customVariables, "custom-variables", "", "Set of infrastructure custom variables.")

	infrastructureCmd.AddCommand(infrastructureDeleteCmd)

	infrastructureCmd.AddCommand(infrastructureDeployCmd)
	infrastructureDeployCmd.Flags().BoolVar(&allowDataLoss, "allow-data-loss", false, "If set, deploy will not throw error if data loss is expected.")
	infrastructureDeployCmd.Flags().BoolVar(&attemptSoftShutdown, "attempt-soft-shutdown", true, "If set, attempt a soft (ACPI) power off of all the servers in the infrastructure before the deploy.")
	infrastructureDeployCmd.Flags().BoolVar(&attemptHardShutdown, "attempt-hard-shutdown", true, "If set, force a hard power off after timeout expired and the server is not powered off")
	infrastructureDeployCmd.Flags().IntVar(&softShutdownTimeout, "soft-shutdown-timeout", 180, "Timeout to wait for soft shutdown before forcing hard shutdown.")
	infrastructureDeployCmd.Flags().BoolVar(&forceShutdown, "force-shutdown", false, "If set, deploy will force shutdown of all servers in the infrastructure.")

	infrastructureCmd.AddCommand(infrastructureRevertCmd)
}
