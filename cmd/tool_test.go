package cmd_test

import (
	"os"
	"path/filepath"

	"github.com/louiss0/go-toolkit/cmd"
	"github.com/louiss0/go-toolkit/internal/modindex/config"
	"github.com/louiss0/go-toolkit/internal/testhelpers"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var Tool = Describe("tool command", func() {
	assert := assert.New(GinkgoT())

	It("installs a tool from the x tools cmd path", func() {
		runner := &testhelpers.RunnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		err := writeDefaultConfig(configPath)
		assert.NoError(err)

		runner.On("Run", mock.Anything, "go", []string{"install", "golang.org/x/tools/cmd/goimports@latest"}).Return(nil).Once()

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
			ConfigPath:   configPath,
		})

		output, err := testhelpers.ExecuteCmd(rootCmd, "tool", "add", "goimports")

		assert.NoError(err)
		assert.Contains(output, "tools added and saved to global packages")
		runner.AssertExpectations(GinkgoT())

		values, err := config.Load(configPath)
		assert.NoError(err)
		assert.Contains(values.GlobalPackages, "golang.org/x/tools/cmd/goimports")
	})

	It("prints the tool install command on dry run", func() {
		runner := &testhelpers.RunnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		err := writeDefaultConfig(configPath)
		assert.NoError(err)

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
			ConfigPath:   configPath,
		})

		output, err := testhelpers.ExecuteCmd(rootCmd, "tool", "add", "goimports", "--dry-run")

		assert.NoError(err)
		runner.AssertNotCalled(GinkgoT(), "Run", mock.Anything, mock.Anything, mock.Anything)
		assert.Contains(output, "go install golang.org/x/tools/cmd/goimports@latest")
	})

	It("uses an explicit slash separated tool path override", func() {
		runner := &testhelpers.RunnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		err := writeDefaultConfig(configPath)
		assert.NoError(err)

		runner.On("Run", mock.Anything, "go", []string{"install", "mvdan.cc/gofumpt@latest"}).Return(nil).Once()

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
			ConfigPath:   configPath,
		})

		_, err = testhelpers.ExecuteCmd(rootCmd, "tool", "add", "mvdan.cc/gofumpt")

		assert.NoError(err)
		runner.AssertExpectations(GinkgoT())

		values, err := config.Load(configPath)
		assert.NoError(err)
		assert.Contains(values.GlobalPackages, "mvdan.cc/gofumpt")
	})

	It("uninstalls a tool from the x tools cmd path", func() {
		runner := &testhelpers.RunnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		err := os.WriteFile(configPath, []byte("user = \"lou\"\nsite = \"github.com\"\nglobal_packages = [\"golang.org/x/tools/cmd/goimports\"]\n"), 0o644)
		assert.NoError(err)

		runner.On("Run", mock.Anything, "go", []string{"clean", "-i", "golang.org/x/tools/cmd/goimports"}).Return(nil).Once()

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
			ConfigPath:   configPath,
		})

		output, err := testhelpers.ExecuteCmd(rootCmd, "tool", "remove", "goimports")

		assert.NoError(err)
		assert.Contains(output, "tools removed from global packages")
		runner.AssertExpectations(GinkgoT())

		values, err := config.Load(configPath)
		assert.NoError(err)
		assert.Empty(values.GlobalPackages)
	})
})
