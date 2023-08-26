package api

const (
	DmPlayerConfig = "dm_player_config" // 更新弹幕配置
	DmPlayer       = "player"

	DmCfg = 1

	// platform
	PlatFormAndroid = "android"
	PlatFormIos     = "ios"
)

var (
	_moduleMap = map[string]int{
		DmPlayerConfig: DmCfg,
		DmPlayer:       DmCfg,
	}
	_configMap = map[int]func(data interface{}) Config{
		DmCfg: NewDmCfg,
	}

	_configModifyMap = map[int]func() ConfigModify{
		DmCfg: NewDmCfgModify,
	}

	_reqToConfigModifyMap = map[int]func(string) ConfigModify{}
)

// VerifyModuleKey verify key
func VerifyModuleKey(key string) int {
	return _moduleMap[key]
}

type ConfigModify interface {
	Merge(req interface{})
	ToMap() (res map[string]string)
}

type Config interface {
	Default()
	Change(req ConfigModify) (modify bool)
	Merge(map[string]string)
	Unmarshal(bs []byte) (err error)
	Marshal() (bs []byte, err error)
}

func NewConfig(moduleId int, data interface{}) Config {
	if op, ok := _configMap[moduleId]; ok {
		return op(data)
	}
	return nil
}

func NewConfigModify(moduleId int) ConfigModify {
	if op, ok := _configModifyMap[moduleId]; ok {
		return op()
	}
	return nil
}

func ReqToConfigModify(req *AddDocReq) ConfigModify {
	var moduleId int
	if req.GetBody() != nil {
		return req.GetBody()
	}
	if moduleId = VerifyModuleKey(req.GetModule()); moduleId == 0 {
		return nil
	}
	if op, ok := _reqToConfigModifyMap[moduleId]; ok {
		return op(req.GetDoc())
	}
	return nil
}
