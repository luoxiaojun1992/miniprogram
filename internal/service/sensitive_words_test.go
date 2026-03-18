package service

import (
	"context"
	"errors"
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

func TestLoadSensitiveWords_NilRepo(t *testing.T) {
	assert.Nil(t, loadSensitiveWords(context.Background(), nil, nil))
}

func TestLoadSensitiveWords_RepoError(t *testing.T) {
	repo := &testutil.MockSensitiveWordRepository{
		ListEnabledWordsFn: func(_ context.Context) ([]string, error) {
			return nil, errors.New("db error")
		},
	}
	assert.Nil(t, loadSensitiveWords(context.Background(), repo, nil))
}

func TestNormalizeSensitiveWords(t *testing.T) {
	got := normalizeSensitiveWords([]string{"  foo  ", "bar", "foo", "", "敏感词"})
	assert.Len(t, got, 3)
	assert.ElementsMatch(t, []string{"敏感词", "foo", "bar"}, got)
}
