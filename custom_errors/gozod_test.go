package custom_errors_test

import (
	"errors"

	"github.com/kaptinlin/gozod"
	"github.com/louiss0/go-toolkit/custom_errors"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("FromZod", func() {
	It("translates themed field validation errors", func() {
		assert := assert.New(GinkgoT())

		schema := gozod.FromStruct[struct {
			ConfigPath string `gozod:"regex=^$|^\\S+$"`
		}]()

		_, err := schema.Parse(struct {
			ConfigPath string `gozod:"regex=^$|^\\S+$"`
		}{ConfigPath: "   "})

		themed := custom_errors.FromZod(err, custom_errors.ZodTheme{
			Subject: "go scaffolding setup",
			FieldMessages: map[string]string{
				"ConfigPath": "config path must not contain spaces",
			},
		})

		assert.Error(err)
		assert.ErrorIs(themed, custom_errors.InvalidInput)
		assert.Contains(themed.Error(), "go scaffolding setup is invalid")
		assert.Contains(themed.Error(), "config path must not contain spaces")
	})

	It("translates root validation errors", func() {
		assert := assert.New(GinkgoT())

		schema := gozod.FromStruct[struct {
			Runner any
		}]().Refine(func(input struct {
			Runner any
		}) bool {
			return input.Runner != nil
		})

		_, err := schema.Parse(struct {
			Runner any
		}{})

		themed := custom_errors.FromZod(err, custom_errors.ZodTheme{
			Subject:     "go scaffolding setup",
			RootMessage: "command wiring requires a runner and prompt runner",
		})

		assert.Error(err)
		assert.Contains(themed.Error(), "command wiring requires a runner and prompt runner")
	})

	It("returns non-zod errors unchanged", func() {
		assert := assert.New(GinkgoT())

		expected := errors.New("plain error")

		actual := custom_errors.FromZod(expected, custom_errors.ZodTheme{
			Subject: "go scaffolding setup",
		})

		assert.Equal(expected, actual)
	})
})
