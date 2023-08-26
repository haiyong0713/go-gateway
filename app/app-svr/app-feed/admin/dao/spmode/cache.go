package spmode

import (
	"context"
	"fmt"
)

func modelUserKey(mid int64) string {
	return fmt.Sprintf("model_user_%d", mid)
}

func devModelUserKey(mobiApp, deviceToken string) string {
	return fmt.Sprintf("device_model_user_%s_%s", mobiApp, deviceToken)
}

func familyRelsOfParentKey(mid int64) string {
	return fmt.Sprintf("fyrel_parent_%d", mid)
}

func familyRelsOfChildKey(mid int64) string {
	return fmt.Sprintf("fyrel_child_%d", mid)
}

//go:generate kratos tool redisgen
type _redis interface {
	// redis: -key=modelUserKey -struct_name=Dao
	DelCacheModelUser(c context.Context, mid int64) error
	// redis: -key=devModelUserKey -struct_name=Dao
	DelCacheDevModelUser(c context.Context, mobiApp string, deviceToken string) error
	// redis: -key=familyRelsOfParentKey -struct_name=Dao
	DelCacheFamilyRelsOfParent(ctx context.Context, id int64) error
	// redis: -key=familyRelsOfChildKey -struct_name=Dao
	DelCacheFamilyRelsOfChild(ctx context.Context, id int64) error
}
