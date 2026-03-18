package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasUnsafeRichText(t *testing.T) {
	t.Run("script tag", func(t *testing.T) {
		assert.True(t, hasUnsafeRichText("<script>alert(1)</script>"))
	})

	t.Run("iframe tag", func(t *testing.T) {
		assert.True(t, hasUnsafeRichText("<iframe src='x'></iframe>"))
	})

	t.Run("javascript url", func(t *testing.T) {
		assert.True(t, hasUnsafeRichText("javascript:alert(1)"))
	})

	t.Run("safe text", func(t *testing.T) {
		assert.False(t, hasUnsafeRichText("normal content"))
	})
}
