package util

const (
	ErrorNet       = "网络错误"
	ErrorDataNull  = "无效ID"
	ErrorTypeFmt   = "错误类型(%s) "
	ErrorPersonFmt = "业务方联系人(%s) "
	ErrorUrlFmt    = "url(%s) "
	ErrorRpcFmt    = "rpc(%s) "
	ErrorErrFmt    = "错误信息(%s) "
	ErrorRes       = ErrorTypeFmt + ErrorPersonFmt + ErrorUrlFmt
	ErrorNetFmts   = ErrorTypeFmt + ErrorUrlFmt + ErrorErrFmt
	ErrorRpcFmts   = "错误类型(Rpc请求错误 %s)" + ErrorPersonFmt + ErrorRpcFmt
	ErrorDBFmts    = "错误类型(DB查询错误 %s)" + ErrorPersonFmt
	ErrorNullFmts  = "错误类型(%s)" + ErrorPersonFmt + ErrorRpcFmt
)
