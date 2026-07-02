package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
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

func NewUninstallCmd(_ runner.Runner, promptRunner prompt.Runner, configPath *string) *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "uninstall [binary] [binaries...]",
		Short: "Remove binaries from GOBIN and clear matching global package entries",
		Args: func(cmd *cobra.Command, args []string) error {
			return validateBinaryInputs(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			values, err := config.Load(*configPath)
			if err != nil {
				return err
			}

			targetBinaries := append([]string{}, args...)
			if len(targetBinaries) == 0 {
				inputs, err := promptBinaryNames(cmd, promptRunner)
				if err != nil {
					if errors.Is(err, huh.ErrUserAborted) {
						return nil
					}
					return err
				}
				targetBinaries = inputs
			}

			if len(targetBinaries) == 0 {
				return custom_errors.CreateInvalidInputErrorWithMessage("at least one binary is required")
			}
			if err := validateBinaryInputs(targetBinaries); err != nil {
				return err
			}

			gobin, err := resolveGoBin()
			if err != nil {
				return err
			}

			binaryPaths := lo.Map(targetBinaries, func(binaryName string, _ int) string {
				return filepath.Join(gobin, executableFileName(binaryName))
			})
			if dryRun {
				return cmdutil.WriteLine(cmd.OutOrStdout(), strings.Join(binaryPaths, "\n"))
			}

			for _, binaryPath := range binaryPaths {
				if err := os.Remove(binaryPath); err != nil {
					return err
				}
			}

			removeSet := lo.SliceToMap(targetBinaries, func(binaryName string) (string, struct{}) {
				return binaryName, struct{}{}
			})
			values.GlobalPackages = lo.Filter(values.GlobalPackages, func(modulePath string, _ int) bool {
				_, shouldRemove := removeSet[moduleBinaryName(modulePath)]
				return !shouldRemove
			})
			if err := config.Save(*configPath, values); err != nil {
				return err
			}

			return cmdutil.WriteLine(cmd.OutOrStdout(), "uninstalled and removed from global packages")
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print the binary paths without deleting them")

	return cmd
}

func promptBinaryNames(cmd *cobra.Command, runner prompt.Runner) ([]string, error) {
	binaryInput, err := runner.Input(cmd, prompt.Input{
		Title:       "Binaries to uninstall",
		Description: "Use space-separated binary names like ginkgo or goimports.",
		Placeholder: "ginkgo goimports",
		Validate: func(value string) error {
			_, err := validation.NonEmptyStrings(strings.Fields(value), "binary names")
			if err != nil {
				return err
			}
			return validateBinaryInputs(strings.Fields(value))
		},
	})
	if err != nil {
		return nil, err
	}

	binaryNames := strings.Fields(binaryInput)
	if err := validateBinaryInputs(binaryNames); err != nil {
		return nil, err
	}

	return binaryNames, nil
}

func validateBinaryInputs(inputs []string) error {
	trimmedInputs, err := validation.NonEmptyStrings(inputs, "binary names")
	if err != nil {
		return err
	}

	if lo.ContainsBy(trimmedInputs, func(input string) bool {
		return !validation.IsBinaryName(input)
	}) {
		return custom_errors.CreateInvalidInputErrorWithMessage("binaries must be bare command names using letters, numbers, underscores, or hyphens")
	}

	return nil
}

func resolveGoBin() (string, error) {
	gobin, ok := os.LookupEnv("GOBIN")
	trimmedGoBin := strings.TrimSpace(gobin)
	if !ok || trimmedGoBin == "" {
		return "", custom_errors.CreateInvalidInputErrorWithMessage("GOBIN must be set to uninstall binaries")
	}

	return trimmedGoBin, nil
}

func executableFileName(binaryName string) string {
	if runtime.GOOS == "windows" {
		return binaryName + ".exe"
	}

	return binaryName
}

func moduleBinaryName(modulePath string) string {
	parts := strings.Split(strings.TrimSpace(modulePath), "/")
	if len(parts) == 0 {
		return ""
	}

	lastPart := parts[len(parts)-1]
	if len(parts) > 1 && versionLikePart(lastPart) {
		return parts[len(parts)-2]
	}

	return lastPart
}

func versionLikePart(value string) bool {
	if len(value) < 2 || value[0] != 'v' {
		return false
	}

	for _, char := range value[1:] {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}
