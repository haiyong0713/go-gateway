package web

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go-gateway/app/web-svr/web-goblin/job/model/web"

	"github.com/pkg/errors"
)

const _xiaomiURI = "/v2/content/post"

func (d *Dao) DelXiaomiArc(ctx context.Context, total int64, bvid string) error {
	nowTs := time.Now().Unix()
	msg := &web.XiaoMiMsg{
		MsgType: 3, //删除
		MsgData: &web.XiaomiMsgData{
			Total: total,
			Articles: []*web.XiaomiArticle{{
				ID: bvid,
			}},
		},
	}
	bs, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	arg := &web.XiaoMiArg{
		AppID:     d.c.Xiaomi.AppID,
		SecretKey: d.xiaomiSign(nowTs),
		Timestamp: nowTs,
		MsgID:     fmt.Sprintf("del_arc_%d", nowTs),
		Msg:       string(bs),
	}
	b := &bytes.Buffer{}
	if err = json.NewEncoder(b).Encode(arg); err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, d.xiaomiURL, b)
	if err != nil {
		return errors.Wrapf(err, "arg:%+v", arg)
	}
	req.Header.Set("Content-Type", "application/json")
	var res struct {
		Code   string `json:"code"`
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	if err = d.xiaomiClient.Do(ctx, req, &res); err != nil {
		return errors.Wrapf(err, "arg:%+v", arg)
	}
	if res.Code != "0" {
		return errors.New(fmt.Sprintf("DelXiaomiArc bvid:%s code:%s status:%s reason:%s", bvid, res.Code, res.Status, res.Reason))
	}
	return nil
}

func (d *Dao) xiaomiSign(ts int64) string {
	md5Str := md5.New()
	_, _ = md5Str.Write([]byte(d.c.Xiaomi.Appkey + strconv.FormatInt(ts, 10)))
	return hex.EncodeToString(md5Str.Sum(nil))
}
