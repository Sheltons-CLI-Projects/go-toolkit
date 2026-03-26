package cmd_test

import (
	"os"
	"path/filepath"

	"github.com/louiss0/go-toolkit/cmd"
	"github.com/louiss0/go-toolkit/internal/testhelpers"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var InstallGlobals = Describe("install-globals command", func() {
	assert := assert.New(GinkgoT())

	It("installs all saved global packages at latest", func() {
		runner := &testhelpers.RunnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		err := os.WriteFile(configPath, []byte("user = \"lou\"\nsite = \"github.com\"\nglobal_packages = [\"github.com/samber/lo\", \"github.com/stretchr/testify\"]\n"), 0o644)
		assert.NoError(err)

		runner.On("Run", mock.Anything, "go", []string{"install", "github.com/samber/lo@latest"}).Return(nil).Once()
		runner.On("Run", mock.Anything, "go", []string{"install", "github.com/stretchr/testify@latest"}).Return(nil).Once()

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
			ConfigPath:   configPath,
		})

		output, err := testhelpers.ExecuteCmd(rootCmd, "install-globals")

		assert.NoError(err)
		assert.Contains(output, "all global packages installed")
		runner.AssertExpectations(GinkgoT())
	})

	It("prints dry run output for install-globals", func() {
		runner := &testhelpers.RunnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		err := os.WriteFile(configPath, []byte("user = \"lou\"\nsite = \"github.com\"\nglobal_packages = [\"github.com/samber/lo\"]\n"), 0o644)
		assert.NoError(err)

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
			ConfigPath:   configPath,
		})

		output, err := testhelpers.ExecuteCmd(rootCmd, "install-globals", "--dry-run")

		assert.NoError(err)
		runner.AssertNotCalled(GinkgoT(), "Run", mock.Anything, mock.Anything, mock.Anything)
		assert.Contains(output, "go install github.com/samber/lo@latest")
	})

	It("errors when no global packages are saved", func() {
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

		_, err = testhelpers.ExecuteCmd(rootCmd, "install-globals")

		assert.Error(err)
		assert.Contains(err.Error(), "no global packages saved")
	})
})
