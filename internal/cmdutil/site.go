package cmdutil

import (
	"github.com/louiss0/go-toolkit/internal/modindex/config"
	"github.com/louiss0/go-toolkit/validation"
	"github.com/spf13/cobra"
)

func ValidateSite(site string, allowFull bool) error {
	return validation.ValidateSite(site, allowFull, config.KnownSites())
}

func RegisterSiteCompletion(cmd *cobra.Command, flagName string) {
	_ = cmd.RegisterFlagCompletionFunc(flagName, func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return config.KnownSites(), cobra.ShellCompDirectiveNoFileComp
	})
}
