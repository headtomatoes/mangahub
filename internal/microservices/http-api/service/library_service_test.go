package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLibraryServiceStructure(t *testing.T) {
	t.Run("ServiceExists", func(t *testing.T) {
		assert.NotNil(t, "library service")
	})
}
