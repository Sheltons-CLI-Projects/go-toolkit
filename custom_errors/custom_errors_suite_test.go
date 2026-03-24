package custom_errors_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCustomErrors(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Custom Errors Suite")
}
