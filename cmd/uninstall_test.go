package cmd_test

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/louiss0/go-toolkit/cmd"
	"github.com/louiss0/go-toolkit/internal/modindex/config"
	"github.com/louiss0/go-toolkit/internal/testhelpers"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var Uninstall = Describe("uninstall command", func() {
	assert := assert.New(GinkgoT())

	It("removes a binary from GOBIN and clears matching global packages", func() {
		runner := &testhelpers.RunnerMock{}
		tempDir := GinkgoT().TempDir()
		gobin := filepath.Join(tempDir, "bin")
		configPath := filepath.Join(tempDir, "config.toml")
		binaryPath := filepath.Join(gobin, binaryFileName("ginkgo"))

		err := os.MkdirAll(gobin, 0o755)
		assert.NoError(err)
		err = os.WriteFile(binaryPath, []byte("binary"), 0o755)
		assert.NoError(err)
		err = os.WriteFile(configPath, []byte("global_packages = [\"github.com/onsi/ginkgo/v2\"]\n"), 0o644)
		assert.NoError(err)

		originalGoBin, hadGoBin := os.LookupEnv("GOBIN")
		assert.NoError(os.Setenv("GOBIN", gobin))
		DeferCleanup(func() {
			if hadGoBin {
				_ = os.Setenv("GOBIN", originalGoBin)
				return
			}
			_ = os.Unsetenv("GOBIN")
		})

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
			ConfigPath:   configPath,
		})

		output, err := testhelpers.ExecuteCmd(rootCmd, "uninstall", "ginkgo")

		assert.NoError(err)
		assert.Contains(output, "uninstalled and removed from global packages")
		runner.AssertNotCalled(GinkgoT(), "Run", mock.Anything, mock.Anything, mock.Anything)

		_, err = os.Stat(binaryPath)
		assert.ErrorIs(err, os.ErrNotExist)

		values, err := config.Load(configPath)
		assert.NoError(err)
		assert.Empty(values.GlobalPackages)
	})

	It("removes a bare binary name like jpd", func() {
		runner := &testhelpers.RunnerMock{}
		tempDir := GinkgoT().TempDir()
		gobin := filepath.Join(tempDir, "bin")
		configPath := filepath.Join(tempDir, "config.toml")
		binaryPath := filepath.Join(gobin, binaryFileName("jpd"))

		assert.NoError(os.MkdirAll(gobin, 0o755))
		assert.NoError(os.WriteFile(binaryPath, []byte("binary"), 0o755))
		assert.NoError(os.WriteFile(configPath, []byte("global_packages = [\"github.com/louiss0/jpd\"]\n"), 0o644))

		originalGoBin, hadGoBin := os.LookupEnv("GOBIN")
		assert.NoError(os.Setenv("GOBIN", gobin))
		DeferCleanup(func() {
			if hadGoBin {
				_ = os.Setenv("GOBIN", originalGoBin)
				return
			}
			_ = os.Unsetenv("GOBIN")
		})

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
			ConfigPath:   configPath,
		})

		_, err := testhelpers.ExecuteCmd(rootCmd, "uninstall", "jpd")
		assert.NoError(err)

		_, err = os.Stat(binaryPath)
		assert.ErrorIs(err, os.ErrNotExist)
	})

	It("prints the binary path on dry run", func() {
		runner := &testhelpers.RunnerMock{}
		tempDir := GinkgoT().TempDir()
		gobin := filepath.Join(tempDir, "bin")
		configPath := filepath.Join(tempDir, "config.toml")

		assert.NoError(os.MkdirAll(gobin, 0o755))
		assert.NoError(writeDefaultConfig(configPath))

		originalGoBin, hadGoBin := os.LookupEnv("GOBIN")
		assert.NoError(os.Setenv("GOBIN", gobin))
		DeferCleanup(func() {
			if hadGoBin {
				_ = os.Setenv("GOBIN", originalGoBin)
				return
			}
			_ = os.Unsetenv("GOBIN")
		})

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
			ConfigPath:   configPath,
		})

		output, err := testhelpers.ExecuteCmd(rootCmd, "uninstall", "ginkgo", "--dry-run")

		assert.NoError(err)
		runner.AssertNotCalled(GinkgoT(), "Run", mock.Anything, mock.Anything, mock.Anything)
		assert.Contains(output, filepath.Join(gobin, binaryFileName("ginkgo")))
	})

	It("rejects package paths and other invalid binary names", func() {
		runner := &testhelpers.RunnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		assert.NoError(writeDefaultConfig(configPath))

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
			ConfigPath:   configPath,
		})

		_, err := testhelpers.ExecuteCmd(rootCmd, "uninstall", "github.com/onsi/ginkgo/v2")
		assert.Error(err)

		_, err = testhelpers.ExecuteCmd(rootCmd, "uninstall", "ginkgo.exe")
		assert.Error(err)
	})

	It("fails when GOBIN is not set", func() {
		runner := &testhelpers.RunnerMock{}
		tempDir := GinkgoT().TempDir()
		configPath := filepath.Join(tempDir, "config.toml")

		assert.NoError(writeDefaultConfig(configPath))

		originalGoBin, hadGoBin := os.LookupEnv("GOBIN")
		_ = os.Unsetenv("GOBIN")
		DeferCleanup(func() {
			if hadGoBin {
				_ = os.Setenv("GOBIN", originalGoBin)
			}
		})

		rootCmd := cmd.NewRootCmdWithOptions(cmd.RootOptions{
			Runner:       runner,
			PromptRunner: testhelpers.NewPromptRunnerMock(),
			ConfigPath:   configPath,
		})

		_, err := testhelpers.ExecuteCmd(rootCmd, "uninstall", "ginkgo")
		assert.Error(err)
	})
})

func binaryFileName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}

	return name
}
