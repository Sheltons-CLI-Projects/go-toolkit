// Package custom_errors provides error handling functionality for flag-related and argument-related operations.
package custom_errors

import (
	"errors"
	"fmt"
	"regexp"
)

// ErrInvalidFlag represents an error indicating an invalid flag.
var ErrInvalidFlag = errors.New("invalid flag")

// ErrInvalidArgument represents an error indicating an invalid argument.
var ErrInvalidArgument = errors.New("invalid argument")

// ErrInvalidInput represents an error indicating invalid user input.
var ErrInvalidInput = errors.New("invalid input")

// FlagName is a string type representing the name of a flag.
type FlagName string

// Error validates the FlagName and returns an error if it's invalid.
// A valid flag name must contain only alphanumeric characters.
func (name FlagName) Error() error {

	regex := regexp.MustCompile(`^[a-z0-9]+$`)

	if !regex.MatchString(string(name)) {
		return fmt.Errorf("%w: %s flag name must be alphanumeric from start to end", ErrInvalidFlag, name)
	}

	return nil
}

// CreateInvalidFlagErrorWithMessage creates an error with a custom message for an invalid flag.
// It first validates the flag name and returns the validation error if present.
var CreateInvalidFlagErrorWithMessage = func(flagName FlagName, message string) error {

	if err := flagName.Error(); err != nil {
		return err
	}

	return fmt.Errorf("%w: %s %s", ErrInvalidFlag, flagName, message)

}

// CreateInvalidArgumentErrorWithMessage creates an error with a custom message for an invalid argument.
var CreateInvalidArgumentErrorWithMessage = func(message string) error {
	return fmt.Errorf("%w: %s", ErrInvalidArgument, message)
}

// CreateInvalidInputErrorWithMessage creates an error with a custom message for invalid input.
var CreateInvalidInputErrorWithMessage = func(message string) error {
	return fmt.Errorf("%w: %s", ErrInvalidInput, message)
}
