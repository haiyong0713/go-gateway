package fawkes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"go-common/library/ecode"

	"go-gateway/app/app-svr/fawkes/service/conf"
	"go-gateway/app/app-svr/fawkes/service/model"
	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"

	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	CometNorify_WXCommonMessage  = "wx_common"
	CometNorify_EmailMessage     = "mail_common"
	CometNorify_PhoneCallMessage = "phone_common"
	CometNorify_PictureMessage   = "wx_picture"
)

// Comet通用消息推送平台 -- 消息推送
func doCometNotify(d *Dao, users []*appmdl.CometUserInfo, notify *appmdl.CometMessageSet, appChannel string) (err error) {
	var (
		req    *http.Request
		reqMdl *appmdl.CometNotifyReq
	)
	reqMdl = &appmdl.CometNotifyReq{App: appChannel, Users: users, Message: notify}
	byteBuf := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(byteBuf)
	encoder.SetEscapeHTML(false)
	if err = encoder.Encode(reqMdl); err != nil {
		log.Error("doCometNotify json encode error(%v)", err)
		return
	}
	if req, err = http.NewRequest(http.MethodPost, d.c.Comet.CometUrl, byteBuf); err != nil {
		log.Error("doCometNotify call http.NewRequest error(%v)", err)
		return
	}
	req.Header.Add("content-type", "application/json")
	req.Header.Add("x-signature", d.c.Comet.Signature)
	req.Header.Add("x-secretid", d.c.Comet.SecretID)
	if err = d.httpClient.Do(context.Background(), req, &reqMdl); err != nil {
		log.Error("doCometNotify error(%v)", reqMdl)
	}
	return
}

func genratorUsers(usernames string) (users []*appmdl.CometUserInfo) {
	usernameList := strings.Split(usernames, "|")
	distinct := sets.NewString(usernameList...).UnsortedList()
	for _, username := range distinct {
		users = append(users, &appmdl.CometUserInfo{Name: username, Phone: ""})
	}
	return
}

// --------------------------------------------------------------------------------------------------------------------

// 企业微信消息推送
func (d *Dao) WechatMessageNotify(messageContent string, usernames string, appChannel string) (error error) {
	var (
		users     []*appmdl.CometUserInfo
		wxMessage *appmdl.CometWeChatMessage
	)
	metadata := appmdl.CometWeChatCardMetadata{}
	wxMessage = &appmdl.CometWeChatMessage{MessageType: CometNorify_WXCommonMessage, Message: messageContent, MessageMetadata: metadata, Template: "", MessageDetail: ""}
	users = genratorUsers(usernames)
	error = doCometNotify(d, users, &appmdl.CometMessageSet{WechatMessage: wxMessage, EmailMessage: nil, PhoneCallMessage: nil}, appChannel)
	return
}

// 企业微信消息卡片推送
func (d *Dao) WechatCardMessageNotify(messageTitle, messageContent, messageLink, messageTemplate string, usernames string, appChannel string) (error error) {
	var (
		users     []*appmdl.CometUserInfo
		wxMessage *appmdl.CometWeChatMessage
	)
	metadata := appmdl.CometWeChatCardMetadata{Subject: messageTitle, Link: messageLink}
	wxMessage = &appmdl.CometWeChatMessage{MessageType: CometNorify_WXCommonMessage, Message: messageContent, MessageMetadata: metadata, Template: messageTemplate, MessageDetail: ""}
	users = genratorUsers(usernames)
	error = doCometNotify(d, users, &appmdl.CometMessageSet{WechatMessage: wxMessage, EmailMessage: nil, PhoneCallMessage: nil}, appChannel)
	return
}

// 邮件消息推送 ( 该服务outlook显示发送人为Notify.暂不支付自定义。 -- 建议使用 mail.go的邮件服务 )
func (d *Dao) EmailMessageNotify(messageTitle, messageContent, messageFrom string, usernames string, appChannel string) (error error) {
	var (
		users        []*appmdl.CometUserInfo
		emailMessage *appmdl.CometEmailMessage
	)
	metadata := appmdl.CometEmailMetadata{Subject: messageTitle, Sender: messageFrom}
	emailMessage = &appmdl.CometEmailMessage{MessageType: CometNorify_EmailMessage, Message: messageContent, MessageMetadata: metadata, Template: ""}
	users = genratorUsers(usernames)
	error = doCometNotify(d, users, &appmdl.CometMessageSet{WechatMessage: nil, EmailMessage: emailMessage, PhoneCallMessage: nil}, d.c.Comet.FawkesAppID)
	return
}

// 电话消息推送
func (d *Dao) PhoneMessageNotify(messageContent string, usernames string) (error error) {
	var (
		users            []*appmdl.CometUserInfo
		phoneCallMessage *appmdl.CometPhoneCallMessage
	)
	phoneCallMessage = &appmdl.CometPhoneCallMessage{MessageType: CometNorify_PhoneCallMessage, Message: messageContent, MessageMetadata: ""}
	users = genratorUsers(usernames)
	error = doCometNotify(d, users, &appmdl.CometMessageSet{WechatMessage: nil, EmailMessage: nil, PhoneCallMessage: phoneCallMessage}, d.c.Comet.FawkesAppID)
	return
}

// 图片消息推送（仅支持企业微信，仅支持无模板的情况）
func (d *Dao) WechatPictureMessageNotify(articles []*appmdl.CometPictureArticleContent, usernames string, appChannel string) (err error) {
	var (
		users          []*appmdl.CometUserInfo
		pictureMessage *appmdl.CometWeChatMessage
	)
	metadata := appmdl.CometPictureMetadata{WxPicture: true}
	message := appmdl.CometPictureArticle{Articles: articles}
	pictureMessage = &appmdl.CometWeChatMessage{MessageType: CometNorify_PictureMessage, Message: message, MessageMetadata: metadata, Template: "", MessageDetail: ""}
	users = genratorUsers(usernames)
	err = doCometNotify(d, users, &appmdl.CometMessageSet{WechatMessage: pictureMessage, EmailMessage: nil, PhoneCallMessage: nil}, appChannel)
	return
}

// EP消息推送
func (d *Dao) WechatEPNotify(content string, usernames string) (err error) {
	var (
		reqURL  string
		req     *http.Request
		data    []byte
		sagaReq *model.SagaReq
		sagaRes *model.SagaRes
	)
	sagatoList := strings.Split(usernames, ",")
	if len(sagatoList) < 1 {
		return
	}
	sagaReq = &model.SagaReq{ToUser: sagatoList, Content: content}
	reqURL = conf.Conf.Host.Saga + "/ep/admin/saga/v2/wechat/message/send"
	if data, err = json.Marshal(sagaReq); err != nil {
		log.Error("WechatEPNotify json marshal error(%v)", err)
		return
	}
	if req, err = http.NewRequest(http.MethodPost, reqURL, strings.NewReader(string(data))); err != nil {
		log.Error("WechatEPNotify call http.NewRequest error(%v)", err)
		return
	}
	req.Header.Add("content-type", "application/json")
	if err = d.httpClient.Do(context.Background(), req, &sagaRes); err != nil {
		log.Error("WechatEPNotify call client.Do error(%v)", err)
		return err
	}
	err = d.ExternalErrorc(context.Background(), int64(ecode.OK.Code()), int64(sagaRes.Code), fmt.Sprintf("WechatEPNotify send, req: %+v\n, resp: %+v\n", sagaReq, sagaRes))
	return
}

// GetWechatToken 获取企微的token
func (d *Dao) GetWechatToken(c context.Context, appSecret string) (token string, err error) {
	// 获取缓存token
	token, err = d.GetCacheWechatToken(c, appSecret)
	if err != nil {
		log.Errorc(c, "GetCacheWechatToken error %v", err)
		return
	}
	if token != "" {
		log.Warnc(c, "get wechat token by redis, token %v", token)
		return
	}
	// 若缓存为空，请求企微接口获取token
	toRequestUrl := fmt.Sprintf(d.c.WXNotify.AccessTokenURL, d.c.WXNotify.CorpID, appSecret)
	req, err := http.NewRequest(http.MethodGet, toRequestUrl, nil)
	if err != nil {
		log.Errorc(c, "http.NewRequest error %v", err)
		return
	}
	req.Header.Add("content-type", "application/json")
	var resp *appmdl.WXNotifyAccessTokenResp
	if err = d.httpClient.Do(context.Background(), req, &resp); err != nil {
		log.Errorc(c, "d.httpClient.Do error %v", err)
		return
	}
	if resp.ErrCode != 0 {
		err = fmt.Errorf("GetWechatToken error,msg %v", resp.ErrMsg)
		return
	}
	token = resp.AccessToken
	log.Warnc(c, "get wechat token by request, token %v", token)
	// token保存在缓存中
	err = d.SetCacheWechatToken(c, d.c.WXNotify.CorpSecret, token, resp.ExpiresIn)
	return
}

// WechatTmpFileUpload 企微临时素材上传
func (d *Dao) WechatTmpFileUpload(c context.Context, fileType string, file multipart.File, fileHeader *multipart.FileHeader, token string) (resp *appmdl.WXNotifyTmpFileResp, err error) {
	toRequestUrl := fmt.Sprintf(d.c.WXNotify.UploadTmpFileURL, token, fileType)
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	part, err := writer.CreateFormFile("media", fileHeader.Filename)
	_, err = io.Copy(part, file)
	if err != nil {
		log.Errorc(c, "io copy error %v", err)
		return
	}
	err = writer.Close()
	if err != nil {
		log.Errorc(c, "writer.close error %v", err)
		return
	}
	req, err := http.NewRequest(http.MethodPost, toRequestUrl, payload)
	if err != nil {
		log.Errorc(c, "http.NewRequest error %v", err)
		return
	}
	req.Header.Set("content-type", writer.FormDataContentType())
	resp = &appmdl.WXNotifyTmpFileResp{}
	if err = d.httpClient.Do(c, req, &resp); err != nil {
		log.Errorc(c, "d.httpClient.Do error %v", err)
	}
	return
}

// WechatAppNotify 企业微信应用消息推送
func (d *Dao) WechatAppNotify(c context.Context, message *appmdl.WXNotifyMessage, token string) (resp *appmdl.WXNotifyMsgResp, err error) {
	var req *http.Request
	// Push Message
	byteBuf := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(byteBuf)
	encoder.SetEscapeHTML(false)
	if err = encoder.Encode(message); err != nil {
		log.Errorc(c, "encoder.Encode %v", err)
		return
	}
	toRequestUrl := fmt.Sprintf(d.c.WXNotify.MessageSendURL, token)
	if req, err = http.NewRequest(http.MethodPost, toRequestUrl, strings.NewReader(byteBuf.String())); err != nil {
		log.Errorc(c, "http.NewRequest error %v", err)
		return
	}
	req.Header.Add("content-type", "application/json")
	if err = d.httpClient.Do(context.Background(), req, &resp); err != nil {
		log.Errorc(c, "d.httpClient.Do error %v", err)
	}
	return
}
