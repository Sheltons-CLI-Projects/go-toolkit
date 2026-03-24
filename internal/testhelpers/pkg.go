package testhelpers

import (
	"bytes"
	"fmt"
	"path/filepath"

	"github.com/louiss0/go-toolkit/internal/prompt"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
)

func CanonicalPath(path string) string {
	resolvedPath, err := filepath.EvalSymlinks(path)
	if err == nil {
		return resolvedPath
	}

	return filepath.Clean(path)
}

type RunnerMock struct {
	mock.Mock
}

func (m *RunnerMock) Run(cmd *cobra.Command, name string, args ...string) error {
	call := m.Called(cmd, name, args)
	return call.Error(0)
}

func ExecuteCmd(cmd *cobra.Command, args ...string) (string, error) {
	buf := new(bytes.Buffer)
	errBuff := new(bytes.Buffer)

	cmd.SetOut(buf)
	cmd.SetErr(errBuff)
	cmd.SetArgs(args)

	err := cmd.Execute()
	if errBuff.Len() > 0 {
		return "", fmt.Errorf("command failed: %s", errBuff.String())
	}

	return buf.String(), err
}

type PromptStepKind string

const (
	PromptStepInput  PromptStepKind = "input"
	PromptStepSelect PromptStepKind = "select"
)

type PromptStep struct {
	Kind  PromptStepKind
	Value string
	Err   error
}

type PromptRunnerMock struct {
	steps []PromptStep
	index int
}

func NewPromptRunnerMock(steps ...PromptStep) *PromptRunnerMock {
	return &PromptRunnerMock{steps: steps}
}

func (m *PromptRunnerMock) Input(_ *cobra.Command, input prompt.Input) (string, error) {
	step, err := m.next(PromptStepInput)
	if err != nil {
		return "", err
	}
	if step.Err != nil {
		return "", step.Err
	}
	if input.Validate != nil {
		if err := input.Validate(step.Value); err != nil {
			return "", err
		}
	}
	return step.Value, nil
}

func (m *PromptRunnerMock) Select(_ *cobra.Command, selectInput prompt.Select) (string, error) {
	step, err := m.next(PromptStepSelect)
	if err != nil {
		return "", err
	}
	if step.Err != nil {
		return "", step.Err
	}
	optionExists := lo.ContainsBy(selectInput.Options, func(option prompt.Option) bool {
		return option.Value == step.Value
	})
	if optionExists {
		return step.Value, nil
	}
	return "", fmt.Errorf("unexpected selection: %s", step.Value)
}

func (m *PromptRunnerMock) next(expected PromptStepKind) (PromptStep, error) {
	if m.index >= len(m.steps) {
		return PromptStep{}, fmt.Errorf("prompt mock: no steps remaining")
	}
	step := m.steps[m.index]
	m.index++
	if step.Kind != expected {
		return PromptStep{}, fmt.Errorf("prompt mock: expected %s, got %s", expected, step.Kind)
	}
	return step, nil
}
