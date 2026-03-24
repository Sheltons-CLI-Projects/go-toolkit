package custom_errors

import (
	"strings"
	"unicode"

	"github.com/kaptinlin/gozod"
	"github.com/samber/lo"
)

type ZodTheme struct {
	Subject       string
	RootMessage   string
	FieldMessages map[string]string
}

func FromZod(err error, theme ZodTheme) error {
	var zodErr *gozod.ZodError
	if !gozod.IsZodError(err, &zodErr) {
		return err
	}

	details := lo.Uniq(lo.FilterMap(zodErr.Issues, func(issue gozod.ZodIssue, _ int) (string, bool) {
		detail := zodIssueDetail(issue, theme)
		return detail, detail != ""
	}))

	if len(details) == 0 {
		return CreateInvalidInputErrorWithMessage(theme.Subject + " is invalid")
	}

	return CreateInvalidInputErrorWithMessage(
		theme.Subject + " is invalid: " + strings.Join(details, "; "),
	)
}

func zodIssueDetail(issue gozod.ZodIssue, theme ZodTheme) string {
	path := gozod.FormatErrorPath(issue.Path, "dot")
	if message, ok := theme.FieldMessages[path]; ok {
		return message
	}

	if path == "" {
		return theme.RootMessage
	}

	return humanizeZodPath(path) + " is invalid"
}

func humanizeZodPath(path string) string {
	var builder strings.Builder

	for index, char := range path {
		if char == '.' || char == '[' || char == ']' {
			builder.WriteRune(' ')
			continue
		}

		if index > 0 && unicode.IsUpper(char) {
			builder.WriteRune(' ')
		}

		builder.WriteRune(unicode.ToLower(char))
	}

	return strings.Join(strings.Fields(builder.String()), " ")
}
