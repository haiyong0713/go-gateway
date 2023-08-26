package http

import (
	"io/ioutil"
	"path"
	"strings"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
)

var (
	_fileTypePCAPm = map[string]struct{}{
		"pcap":   {},
		"pcapng": {},
	}
)

func pcap(ctx *bm.Context) {
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		ctx.JSON(nil, ecode.RequestErr)
		return
	}
	defer file.Close()
	fileName := header.Filename
	fileTpye := strings.TrimPrefix(path.Ext(fileName), ".")
	if _, ok := _fileTypePCAPm[fileTpye]; !ok {
		ctx.String(0, ecode.RequestErr.Error())
		return
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		ctx.String(0, err.Error())
		return
	}
	if len(data) == 0 {
		ctx.String(0, ecode.RequestErr.Error())
		return
	}
	resp, err := svc.Pcap(ctx, fileName, data)
	if err != nil {
		ctx.String(0, err.Error())
		return
	}
	ctx.String(0, resp)
}
