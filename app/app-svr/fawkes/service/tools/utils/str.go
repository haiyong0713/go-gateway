package utils

import "strings"

// Contain return True if list中的某个item与targetStr完全相等或者与targetStr的部分子串相等
func Contain(targetStr string, list []string) bool {
	for _, b := range list {
		if b == targetStr || strings.Contains(targetStr, b) {
			return true
		}
	}
	return false
}
