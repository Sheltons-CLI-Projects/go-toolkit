package cmd

import (
	"errors"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/louiss0/go-toolkit/custom_errors"
	"github.com/louiss0/go-toolkit/internal/cmdutil"
	"github.com/louiss0/go-toolkit/internal/modindex/config"
	"github.com/louiss0/go-toolkit/internal/prompt"
	"github.com/louiss0/go-toolkit/internal/runner"
	"github.com/louiss0/go-toolkit/validation"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

const toolModulePathPrefix = "golang.org/x/tools/cmd"

func NewToolCmd(commandRunner runner.Runner, promptRunner prompt.Runner, configPath *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tool",
		Short: "Install or uninstall Go x/tools executables",
	}

	cmd.AddCommand(
		NewToolAddCmd(commandRunner, promptRunner, configPath),
		NewToolRemoveCmd(commandRunner, promptRunner, configPath),
	)

	return cmd
}

func NewToolAddCmd(commandRunner runner.Runner, promptRunner prompt.Runner, configPath *string) *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "add [tool] [tools...]",
		Short: "Add Go tools from golang.org/x/tools/cmd",
		Args: func(cmd *cobra.Command, args []string) error {
			return validateToolInputs(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			values, err := config.Load(*configPath)
			if err != nil {
				return err
			}

			targetTools := append([]string{}, args...)
			if len(targetTools) == 0 {
				inputs, err := promptToolNames(cmd, promptRunner)
				if err != nil {
					if errors.Is(err, huh.ErrUserAborted) {
						return nil
					}
					return err
				}
				targetTools = inputs
			}

			if len(targetTools) == 0 {
				return custom_errors.CreateInvalidInputErrorWithMessage("at least one tool is required")
			}
			if err := validateToolInputs(targetTools); err != nil {
				return err
			}

			modulePaths := resolveToolModulePaths(targetTools)
			installArgs := lo.Map(modulePaths, func(modulePath string, _ int) string {
				return modulePath + "@latest"
			})

			if dryRun {
				lines := lo.Map(installArgs, func(arg string, _ int) string {
					return "go install " + arg
				})
				return cmdutil.WriteLine(cmd.OutOrStdout(), strings.Join(lines, "\n"))
			}

			for _, arg := range installArgs {
				if err := commandRunner.Run(cmd, "go", "install", arg); err != nil {
					return err
				}
			}

			values.GlobalPackages = lo.Uniq(append(values.GlobalPackages, modulePaths...))
			if err := config.Save(*configPath, values); err != nil {
				return err
			}

			return cmdutil.WriteLine(cmd.OutOrStdout(), "tools added and saved to global packages")
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print the go command without running it")

	return cmd
}

func NewToolRemoveCmd(commandRunner runner.Runner, promptRunner prompt.Runner, configPath *string) *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "remove [tool] [tools...]",
		Short: "Remove Go tools from golang.org/x/tools/cmd",
		Args: func(cmd *cobra.Command, args []string) error {
			return validateToolInputs(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			values, err := config.Load(*configPath)
			if err != nil {
				return err
			}

			targetTools := append([]string{}, args...)
			if len(targetTools) == 0 {
				inputs, err := promptToolNames(cmd, promptRunner)
				if err != nil {
					if errors.Is(err, huh.ErrUserAborted) {
						return nil
					}
					return err
				}
				targetTools = inputs
			}

			if len(targetTools) == 0 {
				return custom_errors.CreateInvalidInputErrorWithMessage("at least one tool is required")
			}
			if err := validateToolInputs(targetTools); err != nil {
				return err
			}

			modulePaths := resolveToolModulePaths(targetTools)
			if dryRun {
				lines := lo.Map(modulePaths, func(modulePath string, _ int) string {
					return "go clean -i " + modulePath
				})
				return cmdutil.WriteLine(cmd.OutOrStdout(), strings.Join(lines, "\n"))
			}

			for _, modulePath := range modulePaths {
				if err := commandRunner.Run(cmd, "go", "clean", "-i", modulePath); err != nil {
					return err
				}
			}

			removeSet := lo.SliceToMap(modulePaths, func(modulePath string) (string, struct{}) {
				return modulePath, struct{}{}
			})
			values.GlobalPackages = lo.Filter(values.GlobalPackages, func(modulePath string, _ int) bool {
				_, shouldRemove := removeSet[modulePath]
				return !shouldRemove
			})
			if err := config.Save(*configPath, values); err != nil {
				return err
			}

			return cmdutil.WriteLine(cmd.OutOrStdout(), "tools removed from global packages")
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print the go command without running it")

	return cmd
}

func promptToolNames(cmd *cobra.Command, runner prompt.Runner) ([]string, error) {
	toolInput, err := runner.Input(cmd, prompt.Input{
		Title:       "Tools",
		Description: "Use tool names like goimports, or override with slash-separated paths.",
		Placeholder: "goimports honnef.co/go/tools/cmd/staticcheck",
		Validate: func(value string) error {
			_, err := validation.RequiredToolList(value, "tools")
			return err
		},
	})
	if err != nil {
		return nil, err
	}

	return validation.RequiredToolList(toolInput, "tools")
}

func validateToolInputs(inputs []string) error {
	trimmedTools, err := validation.NonEmptyStrings(inputs, "tool names")
	if err != nil {
		return err
	}

	if lo.ContainsBy(trimmedTools, func(input string) bool {
		return !validation.IsToolName(input) && !validation.IsToolPath(input)
	}) {
		return custom_errors.CreateInvalidInputErrorWithMessage("tools must be bare names like goimports or slash-separated paths")
	}

	return nil
}

func resolveToolModulePaths(tools []string) []string {
	return lo.Uniq(lo.Map(tools, func(tool string, _ int) string {
		trimmedTool := strings.TrimSpace(tool)
		if validation.IsToolPath(trimmedTool) {
			return trimmedTool
		}

		return toolModulePathPrefix + "/" + trimmedTool
	}))
}
