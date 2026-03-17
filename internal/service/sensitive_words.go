package service

import (
	"context"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/luoxiaojun1992/miniprogram/internal/repository"
)

func loadSensitiveWords(ctx context.Context, repo repository.SensitiveWordRepository, log *logrus.Logger) []string {
	if repo == nil {
		return nil
	}
	words, err := repo.ListEnabledWords(ctx)
	if err != nil {
		if log != nil {
			log.WithError(err).Warn("加载敏感词失败，跳过内容脱敏")
		}
		return nil
	}
	return normalizeSensitiveWords(words)
}

func normalizeSensitiveWords(words []string) []string {
	if len(words) == 0 {
		return nil
	}
	uniq := make(map[string]struct{}, len(words))
	for _, word := range words {
		w := strings.TrimSpace(word)
		if w == "" {
			continue
		}
		uniq[w] = struct{}{}
	}
	if len(uniq) == 0 {
		return nil
	}
	out := make([]string, 0, len(uniq))
	for word := range uniq {
		out = append(out, word)
	}
	sort.Slice(out, func(i, j int) bool {
		return len([]rune(out[i])) > len([]rune(out[j]))
	})
	return out
}

func maskText(text string, words []string) string {
	if text == "" || len(words) == 0 {
		return text
	}
	masked := text
	for _, word := range words {
		runeLen := len([]rune(word))
		if runeLen == 0 {
			continue
		}
		masked = strings.ReplaceAll(masked, word, strings.Repeat("*", runeLen))
	}
	return masked
}
