package lightning

import (
	"errors"
	"testing"

	"github.com/ikidev/lightning/internal/schema"
	"github.com/ikidev/lightning/utils"
)

func TestConversionError(t *testing.T) {
	ok := errors.As(ConversionError{}, &schema.ConversionError{})
	utils.AssertEqual(t, true, ok)
}

func TestUnknownKeyError(t *testing.T) {
	ok := errors.As(UnknownKeyError{}, &schema.UnknownKeyError{})
	utils.AssertEqual(t, true, ok)
}

func TestEmptyFieldError(t *testing.T) {
	ok := errors.As(EmptyFieldError{}, &schema.EmptyFieldError{})
	utils.AssertEqual(t, true, ok)
}

func TestMultiError(t *testing.T) {
	ok := errors.As(MultiError{}, &schema.MultiError{})
	utils.AssertEqual(t, true, ok)
}
