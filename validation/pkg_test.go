package validation_test

import (
	"github.com/louiss0/go-toolkit/validation"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("RequiredString", func() {
	It("returns the trimmed string", func() {
		value, err := validation.RequiredString("  lou  ", "user")

		assert.NoError(err)
		assert.Equal("lou", value)
	})

	It("rejects empty strings", func() {
		_, err := validation.RequiredString("   ", "user")

		assert.Error(err)
		assert.Contains(err.Error(), "user is required")
	})
})

var _ = Describe("ParseBool", func() {
	It("parses boolean strings", func() {
		value, err := validation.ParseBool("true", "enabled")

		assert.NoError(err)
		assert.True(value)
	})

	It("rejects invalid boolean strings", func() {
		_, err := validation.ParseBool("maybe", "enabled")

		assert.Error(err)
		assert.Contains(err.Error(), "enabled must be true or false")
	})
})

var _ = Describe("ValidateSite", func() {
	It("accepts known sites", func() {
		err := validation.ValidateSite("github.com", false, []string{"github.com", "gitlab.com"})

		assert.NoError(err)
	})

	It("rejects unknown sites when full sites are disabled", func() {
		err := validation.ValidateSite("example.com", false, []string{"github.com", "gitlab.com"})

		assert.Error(err)
		assert.Contains(err.Error(), "unsupported site")
	})
})
