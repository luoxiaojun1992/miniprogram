package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/luoxiaojun1992/miniprogram/internal/testutil"
)

func TestMaskText(t *testing.T) {
	repo := &testutil.MockSensitiveWordRepository{
		ListEnabledWordsFn: func(_ context.Context) ([]string, error) {
			return []string{"foo", "敏感词"}, nil
		},
	}
	words := loadSensitiveWords(context.Background(), repo, nil)

	assert.Equal(t, "This is *** and *** content", maskText("This is foo and 敏感词 content", words))
}

func TestMaskText_NoWords(t *testing.T) {
	assert.Equal(t, "normal text", maskText("normal text", nil))
}
