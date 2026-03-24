package cmd

import (
	"errors"
	"path/filepath"

	"github.com/louiss0/go-toolkit/custom_errors"
	"github.com/louiss0/go-toolkit/custom_flags"
	"github.com/louiss0/go-toolkit/internal/cmdutil"
	"github.com/louiss0/go-toolkit/internal/modindex/config"
	"github.com/louiss0/go-toolkit/internal/packagepath"
	"github.com/louiss0/go-toolkit/internal/runner"
	"github.com/louiss0/go-toolkit/internal/scaffold"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

func NewScaffoldCmd(commandRunner runner.Runner, configPath *string) *cobra.Command {
	folderFlag := custom_flags.NewEmptyStringFlag("folder")
	var writeReadme bool
	var initModule bool
	siteFlag := custom_flags.NewEmptyStringFlag("site")
	userFlag := custom_flags.NewEmptyStringFlag("user")
	var allowFull bool

	cmd := &cobra.Command{
		Use:   "scaffold <package name>",
		Short: "Create a package folder with a root file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdutil.LogInfoIfProduction("scaffold: loading config")
			values, err := config.Load(*configPath)
			if err != nil {
				return err
			}

			target := filepath.Clean(args[0])
			packageName := packagepath.NormalizePackageName(filepath.Base(target))

			folder := lo.Ternary(folderFlag.String() == "", target, folderFlag.String())
			writeRootFile := true

			folder = filepath.Clean(folder)
			cmdutil.LogInfoIfProduction("scaffold: creating package at %s", folder)
			if err := scaffold.Create(folder, scaffold.Options{
				PackageName:   packageName,
				WriteRootFile: writeRootFile,
				WriteReadme:   writeReadme,
				WriteTests:    values.Scaffold.WriteTests,
			}); err != nil {
				return err
			}

			if !initModule {
				cmdutil.LogInfoIfProduction("scaffold: module init skipped")
				return nil
			}

			site := config.ResolveSite(siteFlag.String(), values)
			user, err := config.ResolveUser(userFlag.String(), values, site)
			if err != nil {
				if errors.Is(err, config.ErrMissingUser) {
					return custom_errors.CreateInvalidInputErrorWithMessage("missing user; run go-toolkit config set-user <user>")
				}
				return err
			}
			allowCustomSite := allowFull || (siteFlag.String() == "" && values.Site != "")
			if err := cmdutil.ValidateSite(site, allowCustomSite); err != nil {
				return err
			}

			cmdutil.LogInfoIfProduction("scaffold: resolving module path for %s", site)
			modulePath, err := packagepath.ResolveModulePath(packageName, site, user)
			if err != nil {
				return err
			}

			cmdutil.LogInfoIfProduction("scaffold: running go mod init")
			if err := commandRunner.Run(cmd, "go", "-C", folder, "mod", "init", modulePath); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().Var(&folderFlag, "folder", "use a custom folder path")
	cmd.Flags().BoolVar(&writeReadme, "readme", false, "add a README.md to the package")
	cmd.Flags().BoolVar(&initModule, "module", false, "initialize a go.mod for the package")
	cmd.Flags().Var(&userFlag, "user", "override the configured user")
	cmd.Flags().Var(&siteFlag, "site", "override the configured site")
	cmd.Flags().BoolVar(&allowFull, "full", false, "allow a custom module site")
	cmdutil.RegisterSiteCompletion(cmd, "site")

	return cmd
}
