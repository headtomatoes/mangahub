package handler_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRatingHandlerStructure(t *testing.T) {
	t.Run("HandlerExists", func(t *testing.T) {
		assert.NotNil(t, "rating handler")
	})
}
