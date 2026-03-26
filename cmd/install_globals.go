package cmd

import (
	"strings"

	"github.com/louiss0/go-toolkit/custom_errors"
	"github.com/louiss0/go-toolkit/internal/cmdutil"
	"github.com/louiss0/go-toolkit/internal/modindex/config"
	"github.com/louiss0/go-toolkit/internal/runner"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

func NewInstallGlobalsCmd(commandRunner runner.Runner, configPath *string) *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "install-globals",
		Short: "Install all saved global packages at their latest version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			values, err := config.Load(*configPath)
			if err != nil {
				return err
			}

			if len(values.GlobalPackages) == 0 {
				return custom_errors.CreateInvalidInputErrorWithMessage("no global packages saved; use install or config global-package add")
			}

			installArgs := lo.Map(values.GlobalPackages, func(mod string, _ int) string {
				base := strings.Split(mod, "@")[0]
				return base + "@latest"
			})

			if dryRun {
				cmdutil.LogInfoIfProduction("install-globals: dry run output")
				lines := lo.Map(installArgs, func(arg string, _ int) string {
					return "go install " + arg
				})
				return cmdutil.WriteLine(cmd.OutOrStdout(), strings.Join(lines, "\n"))
			}

			cmdutil.LogInfoIfProduction("install-globals: executing go install for %d packages", len(installArgs))
			for _, arg := range installArgs {
				if err := commandRunner.Run(cmd, "go", "install", arg); err != nil {
					return err
				}
			}

			return cmdutil.WriteLine(cmd.OutOrStdout(), "all global packages installed")
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print the go commands without running them")

	return cmd
}
