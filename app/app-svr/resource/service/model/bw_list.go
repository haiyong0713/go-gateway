package model

type BWListWithGroup struct {
	SceneId          int64    //业务场景 ID
	GroupID          int64    //灰度分组 ID
	High             int32    //分组尾号上限
	Low              int32    //分组尾号下限
	GroupToken       string   //分组Token
	IsGroupDeleted   int32    //分组是否失效
	SceneToken       string   //场景token
	DefaultValue     int32    //兜底展现，0-false，1-true，仅在小名单下，对未进名单的oid起效
	LargeOidType     string   //大名单OID类型，目前支持mid和buvid
	LargeListUrl     string   //大名单候选集url
	ListType         string   //业务场景名单类型
	ShowWithoutLogin int32    // 未登录用户是否展现，只在大名单为mid时起效
	SpecialOp        int32    //特殊的灰度切分,0: 按灰度切分，并按接口取值,1:不按灰度切分，仅按接口取值,2: 不按灰度切分，仅按接口取反
	WhiteList        []string //分组白名单
}

const (
	IsDeleted_NORMAL  = int32(0)
	IsDeleted_DELETED = int32(1)

	OidType_OTHER = "other"
	OidType_AVID  = "avid"
	OidType_BVID  = "bvid"
	OidType_MID   = "mid"
	OidType_BUVID = "buvid"

	ListType_SMALL = "small"
	ListType_LARGE = "large"

	SpecialOPBase           = int32(0)
	SpecialOPNoGrayInApi    = int32(1)
	SpecialOPNoGrayNotInApi = int32(2)
)
