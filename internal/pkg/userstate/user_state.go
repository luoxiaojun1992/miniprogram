package userstate

import (
	"strings"

	"github.com/luoxiaojun1992/miniprogram/internal/model/entity"
)

var frozenAttributeNames = map[string]struct{}{
	"is_frozen": {},
}

var mutedAttributeNames = map[string]struct{}{
	"is_muted": {},
}

// IsFrozen checks whether a user should be treated as frozen.
func IsFrozen(attrs []*entity.UserAttribute) bool {
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
