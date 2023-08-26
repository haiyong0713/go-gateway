package tool

import "os"

func IsFileExists(path string) bool {
	_, err := os.Lstat(path)

	return !os.IsNotExist(err)
}

func InInt64Slice(find int64, set []int64) bool {
	for _, v := range set {
		if find == v {
			return true
		}
	}
	return false
}
