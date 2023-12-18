package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		c := NewConfig("", "")
		assert.Equal(t, "8080", c.Port)
		assert.Equal(t, "http://localhost", c.Host)
	})
}
