package testutil

import "testing"

func TestSkipAll(t *testing.T) {
	t.Skip(`No tests in this package for now - This package contains only utility functions that are meant to be used by other functional tests`)
}
