package service

import (
	"os"
	"regexp"
	"strings"
)

func maskSensitiveText(text string) (string, bool) {
	words := loadSensitiveWords()
	if len(words) == 0 || strings.TrimSpace(text) == "" {
		return text, false
	}
	masked := text
	hit := false
	for _, word := range words {
		re := regexp.MustCompile("(?i)" + regexp.QuoteMeta(word))
		if re.MatchString(masked) {
			hit = true
			masked = re.ReplaceAllString(masked, "***")
		}
	}
	return masked, hit
}

func loadSensitiveWords() []string {
	raw := strings.TrimSpace(os.Getenv("APP_SENSITIVE_WORDS"))
	if raw == "" {
		return nil
	}
	raw = strings.NewReplacer("，", ",", "\n", ",", ";", ",", "；", ",").Replace(raw)
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, p := range parts {
		word := strings.TrimSpace(p)
		if word == "" {
			continue
		}
		key := strings.ToLower(word)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, word)
	}
	return out
}

