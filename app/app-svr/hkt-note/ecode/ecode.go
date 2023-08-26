package ecode

import "go-common/library/ecode"

var New = ecode.New

var (
	NoteDBError             = New(79500) // 笔记更新db失败
	NoteOverSizeLimit       = New(79501) // 笔记正文超过上线
	NoteDetailNotFound      = New(79502) // 笔记详情未找到
	NoteContentNotFound     = New(79503) // 笔记正文未找到
	ImageTypeError          = New(79504) // 图片格式有误
	ImageStreamEmpty        = New(79505) // 图片流为空
	ImageTooLarge           = New(79506) // 图片过大
	ImageURLInvalid         = New(79507) // 图片路径错误
	NoteInArcAlreadyExisted = New(79508) // 该稿件已存在笔记，无法新增
	NoteNotFound            = New(79509) // 该笔记不存在或已被删除
	FilterRequestInvalid    = New(79510) // 请求敏感词接口失败
	NoteUserUnfit           = New(79511) // 笔记用户校验未通过
	NoteOverTotalSizeLimit  = New(79512) // 笔记容量超过上线
	NoteOidInvalid          = New(79513) // 笔记所属视频不合法
	ArtDetailNotFound       = New(79514) // 公开笔记详情未找到
	ArtContentNotFound      = New(79515) // 公开笔记正文未找到
	AuthorNotFound          = New(79516) // 笔记作者未找到
	NoteListTypeInvalid     = New(79517) // 笔记列表类型错误
	NoteFrontEndWrong       = New(79518) // 获取html失败
	ArtPublishInvalid       = New(79519) // 笔记不满足发布条件

	ArtNoteCountFail = New(79520) // 获取稿件下笔记数量失败

	TaishanOperationReqInvalid = New(79521) // 泰山操作无效
	TaishanOperationFail       = New(79522) //泰山操作失败

	//todo zdd 加errmsg
	BatchGetReplyRenderInfoFail            = New(79523) //批量获取评论区笔记展示数据失败
	BatchGetReplyRenderInfoReqInvalid      = New(79524) //获取评论区笔记展示数据参数无效
	BatchGetReplyRenderInfoReqCvidOverSize = New(79525) //获取评论区笔记展示数据参数无效,cvid超限
)
