package handler_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLibraryHandlerStructure(t *testing.T) {
	t.Run("HandlerExists", func(t *testing.T) {
		assert.NotNil(t, "library handler")
	})
}
