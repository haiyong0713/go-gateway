package service

import (
	"bytes"
	"context"
	cryRand "crypto/rand"
	"fmt"
	"go-common/library/log"
	"go-common/library/net/trace"
	"go-gateway/app/web-svr/activity/job/conf"
	"math/big"
	"math/rand"
	"time"
)

var shareURLCtx context.Context

func shareURLCtxInit() {
	shareURLCtx = trace.SimpleServerTrace(context.Background(), "collegeUrl")
}

const (
	remainUrl = 1
)

// ShareURLUpdate 分享更新
func (s *Service) ShareURLUpdate() {
	s.shareURLUpdateRunning.Lock()
	defer s.shareURLUpdateRunning.Unlock()
	shareURLCtxInit()
	for b, v := range s.c.Share.ShareLinkConf {
		s.handleSingleFissionShareLink(shareURLCtx, b, v)
	}
	log.Infoc(shareURLCtx, "ShareURLUpdate success()")
}

// handleSingleFissionShareLink 处理单个业务方
func (s *Service) handleSingleFissionShareLink(ctx context.Context, business string, shareConf *conf.SingleShareLinkConf) {
	// 检查当前有效url个
	if shareConf == nil {
		return
	}
	res, err := s.share.ShareURL(ctx, business, shareConf.Token, nil)
	if err != nil {
		return
	}
	var i, resNum int64
	if res != nil {
		resNum = int64(len(res.Location))
	}

	// 生成链接
	removeLinks := make([]string, 0)
	for _, v := range res.Location {
		if int64(len(removeLinks)) < resNum-remainUrl {
			removeLinks = append(removeLinks, v)
		}
	}
	// 移除链接
	s.share.ShareRemoveURL(ctx, business, shareConf.Token, removeLinks)

	// 生成链接
	addLinks := make([]string, 0)
	for i = 0; i < (shareConf.Num - remainUrl); i++ {
		addLinks = append(addLinks, s.createShareLink(shareConf))
	}

	// 添加链接
	s.share.ShareURL(ctx, business, shareConf.Token, addLinks)
}

// createShareLink .
func (s *Service) createShareLink(shareConf *conf.SingleShareLinkConf) string {
	hi := randomInt(len(shareConf.Hosts))
	pi := randomInt(len(shareConf.PrefixPath))
	bp := getRandomString(15)
	return fmt.Sprintf(shareConf.BaseUrl, shareConf.Hosts[hi], shareConf.PrefixPath[pi], bp, time.Now().Unix())
}

// randomInt 随机数
func randomInt(n int) int {
	return rand.Intn(n)
}

// GetRandogetRandomStringmString 获取随机字符串
func getRandomString(len int) string {
	var container string
	var str = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	b := bytes.NewBufferString(str)
	length := b.Len()
	bigInt := big.NewInt(int64(length))
	for i := 0; i < len; i++ {
		randomInt, _ := cryRand.Int(cryRand.Reader, bigInt)
		container += string(str[randomInt.Int64()])
	}
	return container
}
