package userstate

import (
	"strconv"
	"strings"
	"time"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
)

var frozenAttributeNames = map[string]struct{}{
	"is_frozen": {},
}

var mutedAttributeNames = map[string]struct{}{
	"is_muted": {},
}

var mutedUntilAttributeNames = map[string]struct{}{
	"muted_until": {},
}

// IsFrozen checks whether a user should be treated as frozen.
func IsFrozen(attrs []*entity.UserAttribute) bool {
	return hasTruthyAttribute(attrs, frozenAttributeNames)
}

// IsMuted checks whether a user should be treated as muted.
func IsMuted(attrs []*entity.UserAttribute) bool {
	if !hasTruthyAttribute(attrs, mutedAttributeNames) {
		return false
	}
	until := getTimeAttribute(attrs, mutedUntilAttributeNames)
	if until == nil {
		return true
	}
	return time.Now().Before(*until)
}

func hasTruthyAttribute(attrs []*entity.UserAttribute, names map[string]struct{}) bool {
	for _, attr := range attrs {
		if attr == nil || attr.Attribute == nil {
			continue
		}
		name := strings.ToLower(strings.TrimSpace(attr.Attribute.Name))
		if _, ok := names[name]; !ok {
			continue
		}
		if attr.ValueBigint != nil {
			if *attr.ValueBigint == 1 {
				return true
			}
			continue
		}
		if strings.TrimSpace(attr.ValueString) == "1" {
			return true
		}
	}
	return false
}

func getTimeAttribute(attrs []*entity.UserAttribute, names map[string]struct{}) *time.Time {
	for _, attr := range attrs {
		if attr == nil || attr.Attribute == nil {
			continue
		}
		name := strings.ToLower(strings.TrimSpace(attr.Attribute.Name))
		if _, ok := names[name]; !ok {
			continue
		}
		if attr.ValueBigint != nil && *attr.ValueBigint > 0 {
			t := time.Unix(*attr.ValueBigint, 0)
			return &t
		}
		raw := strings.TrimSpace(attr.ValueString)
		if raw == "" {
			continue
		}
		if ts, err := strconv.ParseInt(raw, 10, 64); err == nil && ts > 0 {
			t := time.Unix(ts, 0)
			return &t
		}
		if t, err := time.Parse(time.RFC3339, raw); err == nil {
			return &t
		}
	}
	return nil
}
