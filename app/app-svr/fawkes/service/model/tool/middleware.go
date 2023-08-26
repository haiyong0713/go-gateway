package tool

type key string

var ContentKey key = "ctx_value"

const (
	UNKNOW int8 = iota
	ADMIN
	DEV
	TEST
	DEVOPS
	UNAUTH
	CustomerService
	VISITOR
)

func GetRole(role int8) string {
	switch role {
	case ADMIN:
		return "管理员"
	case DEV:
		return "研发"
	case TEST:
		return "测试"
	case DEVOPS:
		return "运营"
	case UNAUTH:
		return "未授权"
	case CustomerService:
		return "客服"
	case VISITOR:
		return "访客"
	default:
		return "未知"
	}
}
