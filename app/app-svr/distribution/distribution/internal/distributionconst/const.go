package distributionconst

import "fmt"

const (
	PackageNameV1 = "bilibili.app.distribution.v1"
)

const (
	DefaultStorageDriver = "builtin-kv"
)

func MakeFullyQualifiedName(objectName string) string {
	return fmt.Sprintf("%s.%s", PackageNameV1, objectName)
}
