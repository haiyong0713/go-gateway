package tool

import (
	"regexp"
)

var emojiReg = regexp.MustCompile(`\[.*]`)

func RemoveEmojis(input string) string {
	return emojiReg.ReplaceAllString(input, "")
}
