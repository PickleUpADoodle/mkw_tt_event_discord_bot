package commands

import (
	"strings"
)

var _ error = (*MissingArgumentsError)(nil)

type MissingArgumentsError struct {
	Arguments []string
}

func NewMissingArgumentsError(args ...string) error {
	return &MissingArgumentsError{
		Arguments: args,
	}
}

func (e *MissingArgumentsError) Error() string {
	return "Missing arguments: " + strings.Join(e.Arguments, ", ")
}
