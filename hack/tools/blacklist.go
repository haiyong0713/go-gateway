package tools

import "path/filepath"

func InBlacklist(path string) bool {
	return InPathBaseBlacklist(path) || InPathBlacklist(path)
}

var PathBaseBlacklist = map[string]struct{}{
	".idea":     {},
	".git":      {},
	"vendor":    {},
	".vscode":   {},
	".DS_Store": {},
	"_output":   {},
	"ecode":     {},
	"common":    {},
	"test":      {},
}

func InPathBaseBlacklist(path string) bool {
	_, ok := PathBaseBlacklist[filepath.Base(path)]
	return ok
}

var PathBlacklist = map[string]struct{}{}

func InPathBlacklist(path string) bool {
	_, ok := PathBlacklist[path]
	return ok
}
