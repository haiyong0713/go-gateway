package model

import "go-gateway/app/app-svr/archive/service/model"

type CodeType int64

const (
	//default
	CodeType_NOCODE CodeType = 0
	//编码格式 264
	CodeType_CODE264 CodeType = 1
	//编码格式 265
	CodeType_CODE265 CodeType = 2
	// qn hdr
	QnHDR = uint32(125)
	// code H265
	CodeH265 = uint32(12)
	// code H264
	CodeH264 = uint32(7)
	//playurl attribute
	AttrIsHDR = 0
)

func SetQnAttr(attr int64, qn uint32) int64 {
	if qn == model.QnHDR {
		return attr | 1<<model.AttrIsHDR
	}
	return 0
}
