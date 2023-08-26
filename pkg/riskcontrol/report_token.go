package riskcontrol

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"net/url"
	"strings"

	cm "go-common/component/auth/metadata"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"

	"google.golang.org/grpc/metadata"
)

func ReportedLoginTokenFromCtx(ctx context.Context) string {
	if bmCtx, ok := ctx.(*bm.Context); ok {
		accessKey := bmCtx.Request.Form.Get("access_key")
		if accessKey == "" {
			return extractWebReportedToken(bmCtx)
		}
		return md5Hex(accessKey)
	}
	return extractLoginTokenFromGrpc(ctx)
}

func extractLoginTokenFromGrpc(ctx context.Context) string {
	gmd, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Error("extractLoginTokenFromGrpc no parsed gmd")
		return ""
	}
	// authorization: identify_v1 {access_key}
	if authToken := gmd.Get("authorization"); len(authToken) > 0 && authToken[0] != "" && len(authToken[0]) > len("identify_v1 ") {
		var (
			tokenValue = authToken[0]
			idx        = strings.Index(tokenValue, " ")
		)
		if idx == -1 {
			log.Error("extractLoginTokenFromGrpc NoLogin tokenValue=%+v", tokenValue)
			return ""
		}
		tokenType := tokenValue[:idx]
		tokenValue = tokenValue[idx+1:]
		if tokenType != "identify_v1" {
			log.Error("extractLoginTokenFromGrpc authorization failed tokenValue=%+v", tokenValue)
			return ""
		}
		return md5Hex(tokenValue)
	}
	// old authorization
	if clientMeta := gmd.Get("x-bili-metadata-bin"); len(clientMeta) > 0 && len(clientMeta[0]) > 0 {
		var md cm.Metadata
		if err := md.Unmarshal([]byte(clientMeta[0])); err != nil {
			log.Error("extractLoginTokenFromGrpc old authorization failed clientMeta=%+v", clientMeta)
			return ""
		}
		return md5Hex(md.AccessKey)
	}
	log.Error("No authorization field satisfied")
	return ""
}

func extractWebReportedToken(ctx *bm.Context) string {
	req := ctx.Request
	ssDaCk, _ := req.Cookie("SESSDATA")
	if ssDaCk == nil {
		log.Info("extractWebReportedToken no SESSDATA")
		return ""
	}
	ssda, err := url.QueryUnescape(ssDaCk.Value)
	if err != nil {
		log.Error("extractWebReportedToken url.QueryUnescape failed ssDaCk=%+v, err=%+v", ssDaCk, err)
		return ""
	}
	return md5Hex(ssda)
}

func md5Hex(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}
