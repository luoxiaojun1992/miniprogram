package userstate

import (
	"strings"
	"time"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
)

var frozenAttributeNames = map[string]struct{}{
	"freeze":    {},
	"frozen":    {},
	"is_freeze": {},
	"is_frozen": {},
}

var mutedAttributeNames = map[string]struct{}{
	"mute":         {},
	"muted":        {},
	"is_mute":      {},
	"is_muted":     {},
	"comment_mute": {},
	"comment_muted": {},
}

// IsFrozen checks whether a user should be treated as frozen.
func IsFrozen(user *entity.User, attrs []*entity.UserAttribute, now time.Time) bool {
	if user == nil {
		return true
	}
	if user.Status != 1 {
		return true
	}
	if user.FreezeEndTime != nil && now.Before(*user.FreezeEndTime) {
		return true
	}
	return hasTruthyAttribute(attrs, frozenAttributeNames)
}

// IsMuted checks whether a user should be treated as muted.
func IsMuted(attrs []*entity.UserAttribute) bool {
	return hasTruthyAttribute(attrs, mutedAttributeNames)
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
			return *attr.ValueBigint != 0
		}
		switch strings.ToLower(strings.TrimSpace(attr.ValueString)) {
		case "1", "true", "yes", "y", "on":
			return true
		}
	}
	return false
}
