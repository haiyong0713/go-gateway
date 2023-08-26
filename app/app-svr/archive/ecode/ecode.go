package ecode

import (
	"go-common/library/ecode"
)

var (
	// archive
	ArchiveNotExist   = ecode.New(10003) // 不存在该稿件
	VideoshotNotExist = ecode.New(10008) // 稿件的缩略图不存在
)
