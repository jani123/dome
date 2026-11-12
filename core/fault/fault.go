// Package fault supplies stable error categories at package boundaries.
package fault

import (
	"errors"
	"fmt"

	"github.com/samber/oops"
)

// Code is a stable public failure category.
type Code string

const (
	Validation    Code = "validation"
	Observation   Code = "observation"
	Apply         Code = "apply"
	Verification  Code = "verification"
	Persistence   Code = "persistence"
	Conflict      Code = "conflict"
	Configuration Code = "configuration"
	Recovery      Code = "recovery"
)

// Wrap preserves causes; context must contain identifiers only, never secrets.
func Wrap(code Code, err error, context ...any) error {
	if err == nil {
		return nil
	}
	if existing := Category(err); existing != "internal" {
		code = Code(existing)
	}
	return oops.Code(string(code)).With(context...).Wrap(err)
}

// New creates a categorized error; attached context must contain no secrets.
func New(code Code, message string, context ...any) error {
	return Wrap(code, errors.New(message), context...)
}

// Category returns the preserved public error code or internal for uncategorized errors.
func Category(err error) string {
	var e oops.OopsError
	if errors.As(err, &e) {
		if e.Code() != nil && fmt.Sprint(e.Code()) != "" {
			return fmt.Sprint(e.Code())
		}
	}
	return "internal"
}
