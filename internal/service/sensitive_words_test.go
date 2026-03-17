package service

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaskSensitiveText(t *testing.T) {
	old := os.Getenv("APP_SENSITIVE_WORDS")
	t.Cleanup(func() { _ = os.Setenv("APP_SENSITIVE_WORDS", old) })
	_ = os.Setenv("APP_SENSITIVE_WORDS", "foo,敏感词")

	masked, hit := maskSensitiveText("This is foo and 敏感词 content")
	assert.True(t, hit)
	assert.Equal(t, "This is *** and *** content", masked)
}

func TestMaskSensitiveText_NoConfig(t *testing.T) {
	old := os.Getenv("APP_SENSITIVE_WORDS")
	t.Cleanup(func() { _ = os.Setenv("APP_SENSITIVE_WORDS", old) })
	_ = os.Unsetenv("APP_SENSITIVE_WORDS")

	masked, hit := maskSensitiveText("normal text")
	assert.False(t, hit)
	assert.Equal(t, "normal text", masked)
}

